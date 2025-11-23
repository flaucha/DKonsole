// Helper function to get status badge styling (same as nodes in overview)
export const getStatusBadgeClass = (status) => {
    if (!status) return 'bg-gray-700 text-gray-200 border border-gray-600';
    
    const statusLower = status.toLowerCase();
    
    // Green - Ready states
    if (statusLower === 'ready' || statusLower === 'running' || statusLower === 'active' || 
        statusLower === 'available' || statusLower === 'bound' || statusLower === 'succeeded') {
        return 'bg-green-900/50 text-green-300 border border-green-700';
    }
    
    // Red - Error/Failed states
    if (statusLower.includes('error') || statusLower.includes('failed') || 
        statusLower.includes('crashloopbackoff') || statusLower === 'terminating' ||
        statusLower === 'terminated' || statusLower === 'unknown') {
        return 'bg-red-900/50 text-red-300 border border-red-700';
    }
    
    // Orange - Warning states
    if (statusLower === 'pending' || statusLower === 'imagepullbackoff' || 
        statusLower === 'errimagepull' || statusLower.includes('warning')) {
        return 'bg-orange-900/50 text-orange-300 border border-orange-700';
    }
    
    // Yellow - Transitional states
    if (statusLower === 'containercreating' || statusLower === 'init' ||
        statusLower === 'podinitializing' || statusLower.includes('waiting')) {
        return 'bg-yellow-900/50 text-yellow-300 border border-yellow-700';
    }
    
    // Default gray
    return 'bg-gray-700 text-gray-200 border border-gray-600';
};

