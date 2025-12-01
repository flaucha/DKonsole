import { useEffect, useMemo, useState } from 'react';

const sanitizeOrder = (order, availableIds) => {
    const availableSet = new Set(availableIds);
    const filtered = order.filter((id) => availableSet.has(id));
    const missing = availableIds.filter((id) => !filtered.includes(id));
    return [...filtered, ...missing];
};

const readStoredOrder = (storageKey, fallback) => {
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

export const useColumnOrder = (columns, storageKey) => {
    const availableIds = useMemo(() => columns.map((col) => col.id), [columns]);
    const [order, setOrder] = useState(() => readStoredOrder(storageKey, availableIds));

    useEffect(() => {
        if (typeof window === 'undefined') return;
        localStorage.setItem(storageKey, JSON.stringify(order));
    }, [order, storageKey]);

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
