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

describe('useColumnOrder reproduction', () => {
    let cookieMock;

    beforeEach(() => {
        localStorage.clear();
        cookieMock = installCookieMock(';');
    });

    afterEach(() => {
        cookieMock.restore();
    });

    it('should persist order after user login', () => {
        // 1. Initial load without user
        const { result, rerender } = renderHook(
            ({ columns, key, user }) => useColumnOrder(columns, key, user),
            {
                initialProps: { columns: mockColumns, key: 'test-columns', user: null }
            }
        );

        // Verify initial order
        expect(result.current.orderedColumns.map(c => c.id)).toEqual(['name', 'status', 'age']);

        // 2. User logs in
        rerender({ columns: mockColumns, key: 'test-columns', user: 'flaucha' });

        // 3. Reorder columns
        act(() => {
            result.current.moveColumn('name', 'status');
        });

        // Verify order in state
        expect(result.current.orderedColumns.map(c => c.id)).toEqual(['status', 'name', 'age']);

        // Verify cookie
        const cookies = cookieMock.getRaw().split(';');
        const userCookie = cookies.find(c => c.startsWith('test-columns-flaucha='));
        expect(userCookie).toBeTruthy();
        const [, val] = userCookie.split('=');
        expect(decodeURIComponent(val)).toEqual(JSON.stringify(['status', 'name', 'age']));

        // 4. Reload (simulate by unmounting and remounting with user)
        const { result: result2 } = renderHook(
            ({ columns, key, user }) => useColumnOrder(columns, key, user),
            {
                initialProps: { columns: mockColumns, key: 'test-columns', user: 'flaucha' }
            }
        );

        // Verify restored order
        expect(result2.current.orderedColumns.map(c => c.id)).toEqual(['status', 'name', 'age']);
    });
});
