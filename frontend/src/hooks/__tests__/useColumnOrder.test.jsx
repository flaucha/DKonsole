import { renderHook, act } from '@testing-library/react';
import { useColumnOrder } from '../useColumnOrder';

const mockColumns = [
    { id: 'name', label: 'Name' },
    { id: 'status', label: 'Status' },
    { id: 'age', label: 'Age' }
];

const installCookieMock = (separator = '; ') => {
    let store = [];

    Object.defineProperty(document, 'cookie', {
        configurable: true,
        get: () => store.join(separator),
        set: (value) => {
            const [pair] = value.split(';');
            if (!pair) return;
            const [name] = pair.split('=');
            store = store.filter((entry) => !entry.startsWith(`${name}=`));
            store.push(pair);
        }
    });

    return {
        setRaw: (name, value) => {
            const pair = `${name}=${value}`;
            store = store.filter((entry) => !entry.startsWith(`${name}=`));
            store.push(pair);
        },
        getRaw: () => store.join(separator),
        restore: () => {
            store = [];
            delete document.cookie;
        }
    };
};

describe('useColumnOrder cookie persistence', () => {
    let cookieMock;

    beforeEach(() => {
        localStorage.clear();
        cookieMock = installCookieMock(';');
    });

    afterEach(() => {
        cookieMock.restore();
    });

    it('restores order from cookies even when entries have no spaces', () => {
        cookieMock.setRaw('foo', 'bar');
        cookieMock.setRaw('test-columns-alice', JSON.stringify(['status', 'name', 'age']));

        const { result } = renderHook(() => useColumnOrder(mockColumns, 'test-columns', 'alice'));

        expect(result.current.orderedColumns.map((col) => col.id)).toEqual(['status', 'name', 'age']);
    });

    it('persists reordered columns back to cookies', () => {
        const { result } = renderHook(() => useColumnOrder(mockColumns, 'test-columns', 'alice'));

        act(() => {
            result.current.moveColumn('name', 'status');
        });

        const cookies = cookieMock.getRaw().split(';');
        const persisted = cookies.find((entry) => entry.startsWith('test-columns-alice='));
        expect(persisted).toBeTruthy();
        const [, encodedValue] = persisted.split('=');
        const storedOrder = decodeURIComponent(encodedValue);
        expect(storedOrder).toEqual(JSON.stringify(['status', 'name', 'age']));
    });
});
