// Helper function to get status badge styling (same as nodes in overview)
// NO GRAY BADGES - All statuses must be green, red, orange, yellow, or blue
export const getStatusBadgeClass = (status) => {
    if (!status) return 'bg-blue-900/50 text-blue-300 border border-blue-700'; // Unknown status = blue (informative)

    const statusLower = status.toLowerCase();

    // Green - Ready/Healthy states
    if (statusLower === 'ready' || statusLower === 'running' || statusLower === 'available' ||
        statusLower === 'bound' || statusLower === 'succeeded' || statusLower === 'completed' ||
        statusLower === 'healthy' || statusLower === 'ok' || statusLower === 'active') {
        return 'bg-green-900/50 text-green-300 border border-green-700';
    }

    // Red - Error/Failed/Critical states ONLY
    if (statusLower.includes('error') || statusLower.includes('failed') ||
        statusLower.includes('crashloopbackoff') || statusLower === 'terminated' ||
        statusLower.includes('notready') || statusLower.includes('oomkilled') ||
        statusLower.includes('evicted') || statusLower.includes('unhealthy') ||
        statusLower.includes('deadlineexceeded') || statusLower.includes('outofmemory') ||
        statusLower.includes('invalid') || statusLower.includes('rejected') ||
        statusLower === 'unknown') {
        return 'bg-red-900/50 text-red-300 border border-red-700';
    }

    // Yellow - Warning states ONLY
    if (statusLower.includes('warning') || statusLower === 'pending' ||
        statusLower === 'imagepullbackoff' || statusLower === 'errimagepull' ||
        statusLower.includes('schedulingdisabled') || statusLower.includes('unschedulable') ||
        statusLower.includes('degraded') || statusLower.includes('partial') ||
        statusLower === 'terminating' || statusLower === 'suspended') {
        return 'bg-yellow-900/50 text-yellow-300 border border-yellow-700';
    }

    // Blue/Cyan - Informative/Transitional states
    if (statusLower === 'containercreating' || statusLower === 'init' ||
        statusLower === 'podinitializing' || statusLower.includes('waiting') ||
        statusLower.includes('creating') || statusLower.includes('updating') ||
        statusLower.includes('deleting') || statusLower.includes('scaling') ||
        statusLower.includes('provisioning') || statusLower.includes('reconciling')) {
        return 'bg-blue-900/50 text-blue-300 border border-blue-700';
    }

    // Default: Blue for any unknown status (informative, no gray!)
    return 'bg-blue-900/50 text-blue-300 border border-blue-700';
};
