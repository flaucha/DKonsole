import { useEffect, useMemo, useState } from 'react';

const COOKIE_MAX_AGE_SECONDS = 60 * 60 * 24 * 365; // 1 year

const getCookie = (name) => {
    if (typeof document === 'undefined') return null;
    const cookies = document.cookie ? document.cookie.split('; ') : [];
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
    document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; max-age=${COOKIE_MAX_AGE_SECONDS}; path=/; SameSite=Lax`;
};

const storeOrder = (storageKey, order) => {
    if (!storageKey) return;
    try {
        const serialized = JSON.stringify(order);
        setCookie(storageKey, serialized);
        if (typeof window !== 'undefined') {
            localStorage.setItem(storageKey, serialized);
        }
    } catch {
        // Ignore serialization errors
    }
};

const sanitizeOrder = (order, availableIds) => {
    const availableSet = new Set(availableIds);
    const filtered = order.filter((id) => availableSet.has(id));
    const missing = availableIds.filter((id) => !filtered.includes(id));
    return [...filtered, ...missing];
};

const readStoredOrder = (storageKey, fallback) => {
    if (!storageKey) return fallback;
    const storedCookie = getCookie(storageKey);
    if (storedCookie) {
        try {
            const parsedCookie = JSON.parse(storedCookie);
            if (Array.isArray(parsedCookie)) {
                return parsedCookie;
            }
        } catch {
            // Ignore malformed cookies
        }
    }

    if (typeof window === 'undefined') return fallback;

    try {
        const stored = localStorage.getItem(storageKey);
        if (!stored) return fallback;
        const parsed = JSON.parse(stored);
        return Array.isArray(parsed) ? parsed : fallback;
    } catch {
        return fallback;
    }
};

export const useColumnOrder = (columns, storageKey, userId) => {
    const availableIds = useMemo(() => columns.map((col) => col.id), [columns]);
    const storageIdentifier = useMemo(() => {
        if (!storageKey) return null;
        return userId ? `${storageKey}-${userId}` : storageKey;
    }, [storageKey, userId]);
    const [order, setOrder] = useState(() => readStoredOrder(storageIdentifier, availableIds));

    useEffect(() => {
        setOrder(readStoredOrder(storageIdentifier, availableIds));
    }, [storageIdentifier, availableIds]);

    useEffect(() => {
        if (!storageIdentifier) return;
        storeOrder(storageIdentifier, order);
    }, [order, storageIdentifier]);

    useEffect(() => {
        const nextOrder = sanitizeOrder(order, availableIds);
        if (nextOrder.join('|') !== order.join('|')) {
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
        if (storageIdentifier) {
            storeOrder(storageIdentifier, updated);
        }
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
