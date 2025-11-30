/**
 * Permission utilities for checking user permissions
 */

/**
 * Check if user is admin (core admin or LDAP admin)
 * @param {Object} user - User object from auth context
 * @returns {boolean}
 */
export const isAdmin = (user) => {
    if (!user) return false;
    return user.role === 'admin';
};

/**
 * Check if user is LDAP admin
 * @param {Object} user - User object from auth context
 * @returns {boolean}
 */
export const isLDAPAdmin = (user) => {
    if (!user) return false;
    return user.role === 'admin' && user.idp === 'ldap';
};

/**
 * Check if user is core admin
 * @param {Object} user - User object from auth context
 * @returns {boolean}
 */
export const isCoreAdmin = (user) => {
    if (!user) return false;
    // Only return true if explicitly core admin (idp === 'core')
    // If idp is undefined, we can't be sure, so return false for safety
    return user.role === 'admin' && user.idp === 'core';
};

/**
 * Check if user has edit permission for a namespace
 * @param {Object} user - User object from auth context
 * @param {string} namespace - Namespace to check
 * @returns {boolean}
 */
export const hasEditPermission = (user, namespace) => {
    if (!user || !namespace) return false;
    if (isAdmin(user)) return true;
    if (!user.permissions) return false;
    const permission = user.permissions[namespace];
    return permission === 'edit';
};

/**
 * Check if user has view permission for a namespace
 * @param {Object} user - User object from auth context
 * @param {string} namespace - Namespace to check
 * @returns {boolean}
 */
export const hasViewPermission = (user, namespace) => {
    if (!user || !namespace) return false;
    if (isAdmin(user)) return true;
    if (!user.permissions) return false;
    const permission = user.permissions[namespace];
    return permission === 'view' || permission === 'edit';
};

/**
 * Check if user can edit (has edit permission or is admin)
 * @param {Object} user - User object from auth context
 * @param {string} namespace - Namespace to check
 * @returns {boolean}
 */
export const canEdit = (user, namespace) => {
    if (!user || !namespace) return false;
    if (isAdmin(user)) return true;
    return hasEditPermission(user, namespace);
};

/**
 * Check if user can view (has view or edit permission or is admin)
 * @param {Object} user - User object from auth context
 * @param {string} namespace - Namespace to check
 * @returns {boolean}
 */
export const canView = (user, namespace) => {
    if (!user || !namespace) return false;
    if (isAdmin(user)) return true;
    return hasViewPermission(user, namespace);
};
