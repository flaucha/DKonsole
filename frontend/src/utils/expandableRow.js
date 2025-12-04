/**
 * Standard styles and utilities for expandable rows across the application
 */

/**
 * Get the standard classes for an expandable row container
 * @param {boolean} isExpanded - Whether the row is currently expanded
 * @param {boolean} hasPadding - Whether to include left padding (default: true)
 * @returns {string} CSS classes for the expandable container
 */
export const getExpandableRowClasses = (isExpanded, hasPadding = true) => {
    const baseClasses = 'transition-all duration-300 ease-in-out';
    const paddingClass = hasPadding ? 'pl-12' : '';
    const stateClasses = isExpanded 
        ? 'opacity-100 max-h-[80vh] overflow-y-auto' 
        : 'max-h-0 opacity-0 overflow-hidden';
    
    return `${baseClasses} ${paddingClass} ${stateClasses}`.trim();
};

/**
 * Get the standard classes for an expandable row's table cell
 * @param {boolean} isExpanded - Whether the row is currently expanded
 * @param {number} colSpan - Number of columns to span
 * @returns {string} CSS classes for the table cell
 */
export const getExpandableCellClasses = (isExpanded) => {
    return `px-6 pt-0 bg-gray-800 border-0 ${isExpanded ? 'border-b border-gray-700' : ''}`;
};

/**
 * Get the standard inline styles for expandable rows that need max-height
 * @param {boolean} isExpanded - Whether the row is currently expanded
 * @param {string} kind - Resource kind (e.g., 'Pod')
 * @param {string} customMaxHeight - Custom max height (optional)
 * @returns {object} Inline styles object
 */
export const getExpandableRowStyles = (isExpanded, kind = null, customMaxHeight = null) => {
    if (!isExpanded) return {};
    
    if (customMaxHeight) {
        return { maxHeight: customMaxHeight };
    }
    
    // Special handling for Pods which need more height
    if (kind === 'Pod') {
        return { maxHeight: 'calc(100vh - 250px)' };
    }
    
    return {};
};

/**
 * Get the standard classes for a row that can be expanded
 * @param {boolean} isExpanded - Whether the row is currently expanded
 * @returns {string} CSS classes for the expandable row
 */
export const getExpandableRowRowClasses = (isExpanded) => {
    return `group hover:bg-gray-800/50 transition-colors cursor-pointer ${isExpanded ? 'bg-gray-800/30' : ''}`;
};




