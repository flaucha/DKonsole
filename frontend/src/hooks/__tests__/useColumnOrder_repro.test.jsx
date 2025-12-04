import { renderHook } from '@testing-library/react';
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

    it('should not overwrite new key with old order when switching keys', () => {
        // 1. Setup: Key A has specific order
        cookieMock.setRaw('test-columns-A', JSON.stringify(['age', 'name', 'status']));

        const { result, rerender } = renderHook(
            ({ columns, key, user }) => useColumnOrder(columns, key, user),
            {
                initialProps: { columns: mockColumns, key: 'test-columns-A', user: null }
            }
        );

        expect(result.current.orderedColumns.map(c => c.id)).toEqual(['age', 'name', 'status']);

        // 2. Switch to Key B (which has NO saved order, so should be default)
        rerender({ columns: mockColumns, key: 'test-columns-B', user: null });

        // 3. Verify Key B order is default
        expect(result.current.orderedColumns.map(c => c.id)).toEqual(['name', 'status', 'age']);

        // 4. Verify Key B was NOT saved with Key A's order
        const cookies = cookieMock.getRaw().split(';');
        const keyBCookie = cookies.find(c => c.startsWith('test-columns-B='));

        if (keyBCookie) {
            const [, val] = keyBCookie.split('=');
            expect(decodeURIComponent(val)).not.toEqual(JSON.stringify(['age', 'name', 'status']));
        }
    });
});
