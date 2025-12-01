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

const storeOrder = (storageKeys, data) => {
    const keys = Array.isArray(storageKeys) ? storageKeys : [storageKeys];
    keys.filter(Boolean).forEach((key) => {
        try {
            const serialized = JSON.stringify(data);
            setCookie(key, serialized);
            if (typeof window !== 'undefined') {
                localStorage.setItem(key, serialized);
            }
        } catch (err) {
            console.error(`[useColumnOrder] Failed to save data for key ${key}:`, err);
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
    console.debug(`[useColumnOrder] Reading data from keys:`, keys);

    for (const storageKey of keys) {
        if (!storageKey) continue;

        const storedCookie = getCookie(storageKey);
        if (storedCookie) {
            try {
                const parsed = JSON.parse(storedCookie);
                // Backward compatibility: if array, it's just order
                if (Array.isArray(parsed)) {
                    console.debug(`[useColumnOrder] Found legacy order in cookie for key ${storageKey}:`, parsed);
                    return { order: parsed, hidden: [] };
                }
                // New format: object with order and hidden
                if (parsed && Array.isArray(parsed.order)) {
                    console.debug(`[useColumnOrder] Found data in cookie for key ${storageKey}:`, parsed);
                    return { order: parsed.order, hidden: parsed.hidden || [] };
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
                        console.debug(`[useColumnOrder] Found legacy order in localStorage for key ${storageKey}:`, parsed);
                        return { order: parsed, hidden: [] };
                    }
                    if (parsed && Array.isArray(parsed.order)) {
                        console.debug(`[useColumnOrder] Found data in localStorage for key ${storageKey}:`, parsed);
                        return { order: parsed.order, hidden: parsed.hidden || [] };
                    }
                }
            } catch (err) {
                console.warn(`[useColumnOrder] Malformed localStorage for key ${storageKey}:`, err);
            }
        }
    }
    console.debug(`[useColumnOrder] No stored data found, using fallback.`);
    return { order: fallback, hidden: [] };
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

    const [state, setState] = useState(() => {
        const stored = readStoredOrder(keysToRead, availableIds);
        return {
            key: keysToRead.join(','),
            order: stored.order,
            hidden: stored.hidden
        };
    });

    // Derived state pattern: if keys change, update state immediately during render
    const currentKeyStr = keysToRead.join(',');
    if (state.key !== currentKeyStr) {
        console.debug(`[useColumnOrder] Keys changed from ${state.key} to ${currentKeyStr}, reloading data synchronously.`);
        const newStored = readStoredOrder(keysToRead, availableIds);
        setState({
            key: currentKeyStr,
            order: newStored.order,
            hidden: newStored.hidden
        });
    }

    const order = state.order;
    const hidden = state.hidden;

    const setOrder = (newOrder) => {
        setState(prev => ({ ...prev, order: newOrder }));
    };

    const toggleVisibility = (columnId) => {
        setState(prev => {
            const isHidden = prev.hidden.includes(columnId);
            const newHidden = isHidden
                ? prev.hidden.filter(id => id !== columnId)
                : [...prev.hidden, columnId];
            return { ...prev, hidden: newHidden };
        });
    };

    useEffect(() => {
        // Only save if the key in state matches the current key
        if (state.key === keysToWrite.join(',')) {
            storeOrder(keysToWrite, { order, hidden });
        }
    }, [order, hidden, keysToWrite, state.key]);

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
    };

    const orderedColumns = useMemo(() => {
        const orderIndex = sanitizeOrder(order, availableIds).reduce((acc, id, idx) => {
            acc[id] = idx;
            return acc;
        }, {});
        return [...columns].sort((a, b) => (orderIndex[a.id] ?? columns.length) - (orderIndex[b.id] ?? columns.length));
    }, [columns, order, availableIds]);

    const visibleColumns = useMemo(() => {
        return orderedColumns.filter(col => !hidden.includes(col.id));
    }, [orderedColumns, hidden]);

    const resetOrder = () => setState(prev => ({ ...prev, order: availableIds, hidden: [] }));

    return {
        orderedColumns, // All columns, ordered
        visibleColumns, // Only visible columns, ordered
        moveColumn,
        resetOrder,
        order,
        hidden,
        toggleVisibility
    };
};

export default useColumnOrder;
