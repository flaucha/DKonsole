/**
 * Standard date formatting utilities for consistent date display across the application
 */

/**
 * Format a date string to a standard date-time format
 * Format: MM/DD/YYYY, HH:MM:SS (en-US locale)
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Formatted date string or 'Unknown' if invalid
 */
export const formatDateTime = (dateString) => {
    if (!dateString) return 'Unknown';
    try {
        return new Date(dateString).toLocaleString('en-US', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit'
        });
    } catch {
        return 'Unknown';
    }
};

/**
 * Format a date string to a standard date-time format without seconds
 * Format: MM/DD/YYYY, HH:MM (en-US locale)
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Formatted date string or 'Unknown' if invalid
 */
export const formatDateTimeShort = (dateString) => {
    if (!dateString) return 'Unknown';
    try {
        return new Date(dateString).toLocaleString('en-US', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    } catch {
        return 'Unknown';
    }
};

/**
 * Format a date string to date only (no time)
 * Format: MM/DD/YYYY (en-US locale)
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Formatted date string or 'Unknown' if invalid
 */
export const formatDate = (dateString) => {
    if (!dateString) return 'Unknown';
    try {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit'
        });
    } catch {
        return 'Unknown';
    }
};

/**
 * Format a date string to a relative time (e.g., "2h ago", "3d ago")
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Relative time string or 'Unknown' if invalid
 */
export const formatRelativeTime = (dateString) => {
    if (!dateString) return 'Unknown';
    try {
        const diff = Date.now() - new Date(dateString).getTime();
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        if (days > 0) return `${days}d ago`;
        const hours = Math.floor(diff / (1000 * 60 * 60));
        if (hours > 0) return `${hours}h ago`;
        const minutes = Math.floor(diff / (1000 * 60));
        if (minutes > 0) return `${minutes}m ago`;
        return 'Just now';
    } catch {
        return 'Unknown';
    }
};

/**
 * Format a date string to age format (e.g., "2d", "3h", "45m")
 * @param {string|Date} dateString - ISO date string or Date object
 * @returns {string} Age string or 'Unknown' if invalid
 */
export const formatAge = (dateString) => {
    if (!dateString) return 'Unknown';
    try {
        const diff = Date.now() - new Date(dateString).getTime();
        const days = Math.floor(diff / (1000 * 60 * 60 * 24));
        if (days > 0) return `${days}d`;
        const hours = Math.floor(diff / (1000 * 60 * 60));
        if (hours > 0) return `${hours}h`;
        const minutes = Math.floor(diff / (1000 * 60));
        return `${minutes}m`;
    } catch {
        return 'Unknown';
    }
};



