import { useEffect, useMemo, useState } from 'react';

const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365; // 1 year

const getCookie = (name) => {
    if (typeof document === 'undefined' || !name) return null;
    const raw = document.cookie || '';
    if (!raw) return null;
    const cookies = raw.split(';').map((cookie) => cookie.trim()).filter(Boolean);
    for (const cookie of cookies) {
        const [key, ...valueParts] = cookie.split('=');
        if (key === name) {
            return valueParts.length > 0 ? decodeURIComponent(valueParts.join('=')) : '';
        }
    }
    return null;
};

const setCookie = (name, value) => {
    if (typeof document === 'undefined' || !name) return;
    const expires = new Date(Date.now() + COOKIE_MAX_AGE_SECONDS * 1000).toUTCString();
    // Removing SameSite=Lax to ensure compatibility with .lan domains and potentially non-secure contexts
    // Modern browsers default to Lax, but explicit setting might be interfering if not on HTTPS
    document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; max-age=${COOKIE_MAX_AGE_SECONDS}; path=/`;
    console.debug(`[useColumnOrder] Saved cookie: ${name}`, value);
};

const storeOrder = (storageKeys, order) => {
    const keys = Array.isArray(storageKeys) ? storageKeys : [storageKeys];
    keys.filter(Boolean).forEach((key) => {
        try {
            const serialized = JSON.stringify(order);
            setCookie(key, serialized);
            if (typeof window !== 'undefined') {
                localStorage.setItem(key, serialized);
            }
        } catch (err) {
            console.error(`[useColumnOrder] Failed to save order for key ${key}:`, err);
        }
    });
};

const sanitizeOrder = (order, availableIds) => {
    const availableSet = new Set(availableIds);
    const filtered = order.filter((id) => availableSet.has(id));
    const missing = availableIds.filter((id) => !filtered.includes(id));
    return [...filtered, ...missing];
};

const readStoredOrder = (storageKeys, fallback) => {
    const keys = Array.isArray(storageKeys) ? storageKeys : [storageKeys];
    console.debug(`[useColumnOrder] Reading order from keys:`, keys);

    for (const storageKey of keys) {
        if (!storageKey) continue;

        const storedCookie = getCookie(storageKey);
        if (storedCookie) {
            try {
                const parsedCookie = JSON.parse(storedCookie);
                if (Array.isArray(parsedCookie)) {
                    console.debug(`[useColumnOrder] Found order in cookie for key ${storageKey}:`, parsedCookie);
                    return parsedCookie;
                }
            } catch (err) {
                console.warn(`[useColumnOrder] Malformed cookie for key ${storageKey}:`, err);
            }
        }

        if (typeof window !== 'undefined') {
            try {
                const stored = localStorage.getItem(storageKey);
                if (stored) {
                    const parsed = JSON.parse(stored);
                    if (Array.isArray(parsed)) {
                        console.debug(`[useColumnOrder] Found order in localStorage for key ${storageKey}:`, parsed);
                        return parsed;
                    }
                }
            } catch (err) {
                console.warn(`[useColumnOrder] Malformed localStorage for key ${storageKey}:`, err);
            }
        }
    }
    console.debug(`[useColumnOrder] No stored order found, using fallback.`);
    return fallback;
};

export const useColumnOrder = (columns, storageKey, userId) => {
    const availableIds = useMemo(() => columns.map((col) => col.id), [columns]);
    const baseKey = useMemo(() => storageKey || null, [storageKey]);
    const userKey = useMemo(() => {
        if (!storageKey) return null;
        return userId ? `${storageKey}-${userId}` : null;
    }, [storageKey, userId]);
    const keysToRead = useMemo(() => (userKey ? [userKey, baseKey] : [baseKey]), [userKey, baseKey]);
    const keysToWrite = useMemo(() => (userKey ? [userKey, baseKey] : [baseKey]), [userKey, baseKey]);

    const [state, setState] = useState(() => ({
        key: keysToRead.join(','),
        order: readStoredOrder(keysToRead, availableIds)
    }));

    // Derived state pattern: if keys change, update state immediately during render
    // This prevents the "old order, new key" race condition in useEffect
    const currentKeyStr = keysToRead.join(',');
    if (state.key !== currentKeyStr) {
        console.debug(`[useColumnOrder] Keys changed from ${state.key} to ${currentKeyStr}, reloading order synchronously.`);
        const newOrder = readStoredOrder(keysToRead, availableIds);
        setState({
            key: currentKeyStr,
            order: newOrder
        });
    }

    const order = state.order;
    const setOrder = (newOrder) => {
        setState(prev => ({ ...prev, order: newOrder }));
    };

    // Removed the useEffect that loaded order on key change, as it's now handled synchronously above.

    useEffect(() => {
        // Only save if the key in state matches the current key (double safety)
        if (state.key === keysToWrite.join(',')) {
            storeOrder(keysToWrite, order);
        }
    }, [order, keysToWrite, state.key]);

    useEffect(() => {
        const nextOrder = sanitizeOrder(order, availableIds);
        if (nextOrder.join('|') !== order.join('|')) {
            console.debug(`[useColumnOrder] Sanitizing order (columns changed).`);
            setOrder(nextOrder);
        }
    }, [order, availableIds]);

    const moveColumn = (sourceId, targetId) => {
        if (sourceId === targetId) return;
        const sourceIndex = order.indexOf(sourceId);
        const targetIndex = order.indexOf(targetId);
        if (sourceIndex === -1 || targetIndex === -1) return;
        const updated = [...order];
        updated.splice(sourceIndex, 1);
        updated.splice(targetIndex, 0, sourceId);
        setOrder(updated);
        // We can save immediately here too, but the effect handles it.
    };

    const orderedColumns = useMemo(() => {
        const orderIndex = sanitizeOrder(order, availableIds).reduce((acc, id, idx) => {
            acc[id] = idx;
            return acc;
        }, {});
        return [...columns].sort((a, b) => (orderIndex[a.id] ?? columns.length) - (orderIndex[b.id] ?? columns.length));
    }, [columns, order, availableIds]);

    const resetOrder = () => setOrder(availableIds);

    return {
        orderedColumns,
        moveColumn,
        resetOrder,
        order
    };
};

export default useColumnOrder;
