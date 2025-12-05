import { describe, it, expect } from 'vitest';
import {
    isAdmin,
    isLDAPAdmin,
    isCoreAdmin,
    hasEditPermission,
    hasViewPermission,
    canEdit,
    canView
} from '../permissions';

describe('permissions util', () => {

    describe('isAdmin', () => {
        it('should return false for null/undefined user', () => {
            expect(isAdmin(null)).toBe(false);
            expect(isAdmin(undefined)).toBe(false);
        });

        it('should return true if user role is admin', () => {
            expect(isAdmin({ role: 'admin' })).toBe(true);
        });

        it('should return false if user role is not admin', () => {
            expect(isAdmin({ role: 'user' })).toBe(false);
            expect(isAdmin({ role: '' })).toBe(false);
        });
    });

    describe('isLDAPAdmin', () => {
        it('should return false for null/undefined user', () => {
            expect(isLDAPAdmin(null)).toBe(false);
        });

        it('should return true only if user is admin AND idp is ldap', () => {
            expect(isLDAPAdmin({ role: 'admin', idp: 'ldap' })).toBe(true);
        });

        it('should return false if user is admin but not ldap', () => {
            expect(isLDAPAdmin({ role: 'admin', idp: 'core' })).toBe(false);
            expect(isLDAPAdmin({ role: 'admin' })).toBe(false);
        });

        it('should return false if user is ldap but not admin', () => {
            expect(isLDAPAdmin({ role: 'user', idp: 'ldap' })).toBe(false);
        });
    });

    describe('isCoreAdmin', () => {
        it('should return false for null user', () => {
            expect(isCoreAdmin(null)).toBe(false);
        });

        it('should return true only if user is admin AND idp is core', () => {
            expect(isCoreAdmin({ role: 'admin', idp: 'core' })).toBe(true);
        });

        it('should return false if user is admin but not core', () => {
            expect(isCoreAdmin({ role: 'admin', idp: 'ldap' })).toBe(false);
            expect(isCoreAdmin({ role: 'admin' })).toBe(false);
        });
    });

    describe('hasEditPermission', () => {
        it('should return false for invalid inputs', () => {
            expect(hasEditPermission(null, 'default')).toBe(false);
            expect(hasEditPermission({}, null)).toBe(false);
        });

        it('should return true if user is admin', () => {
            expect(hasEditPermission({ role: 'admin' }, 'any-ns')).toBe(true);
        });

        it('should return false if user has no permissions object', () => {
            expect(hasEditPermission({ role: 'user' }, 'default')).toBe(false);
        });

        it('should return true if user has edit permission for namespace', () => {
            const user = { role: 'user', permissions: { 'default': 'edit' } };
            expect(hasEditPermission(user, 'default')).toBe(true);
        });

        it('should return false if user has view permission for namespace', () => {
            const user = { role: 'user', permissions: { 'default': 'view' } };
            expect(hasEditPermission(user, 'default')).toBe(false);
        });

        it('should return false if namespace permission is missing', () => {
            const user = { role: 'user', permissions: { 'other': 'edit' } };
            expect(hasEditPermission(user, 'default')).toBe(false);
        });
    });

    describe('hasViewPermission', () => {
        it('should return false for invalid inputs', () => {
            expect(hasViewPermission(null, 'default')).toBe(false);
        });

        it('should return true if user is admin', () => {
            expect(hasViewPermission({ role: 'admin' }, 'any-ns')).toBe(true);
        });

        it('should return true if user has view permission', () => {
            const user = { role: 'user', permissions: { 'default': 'view' } };
            expect(hasViewPermission(user, 'default')).toBe(true);
        });

        it('should return true if user has edit permission (implies view)', () => {
            const user = { role: 'user', permissions: { 'default': 'edit' } };
            expect(hasViewPermission(user, 'default')).toBe(true);
        });

        it('should return false if no permission for namespace', () => {
            const user = { role: 'user', permissions: { 'other': 'edit' } };
            expect(hasViewPermission(user, 'default')).toBe(false);
        });
    });

    describe('canEdit', () => {
        it('should return true if admin', () => {
            expect(canEdit({ role: 'admin' }, 'ns')).toBe(true);
        });
        it('should return true if has edit perm', () => {
            const user = { role: 'user', permissions: { 'ns': 'edit' } };
            expect(canEdit(user, 'ns')).toBe(true);
        });
        it('should return false if view perm only', () => {
            const user = { role: 'user', permissions: { 'ns': 'view' } };
            expect(canEdit(user, 'ns')).toBe(false);
        });
    });

    describe('canView', () => {
        it('should return true if admin', () => {
            expect(canView({ role: 'admin' }, 'ns')).toBe(true);
        });
        it('should return true if has view perm', () => {
            const user = { role: 'user', permissions: { 'ns': 'view' } };
            expect(canView(user, 'ns')).toBe(true);
        });
        it('should return true if has edit perm', () => {
            const user = { role: 'user', permissions: { 'ns': 'edit' } };
            expect(canView(user, 'ns')).toBe(true);
        });
        it('should return false if no perm', () => {
            const user = { role: 'user', permissions: {} };
            expect(canView(user, 'ns')).toBe(false);
        });
    });

});
