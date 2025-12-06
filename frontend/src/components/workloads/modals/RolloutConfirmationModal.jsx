import React from 'react';

const RolloutConfirmationModal = ({ confirmRollout, setConfirmRollout, handleRolloutDeployment }) => {
    if (!confirmRollout) return null;

    const details = confirmRollout.details;
    const strategy = confirmRollout.strategy;
    const replicas = details?.replicas || 0;
    const ready = details?.ready || 0;
    const readyCount = typeof ready === 'string' ? parseInt(ready.split('/')[0]) || 0 : ready;
    const totalReplicas = typeof replicas === 'string' ? parseInt(replicas.split('/')[1]) || parseInt(replicas) || 0 : replicas;

    // Parse maxSurge and maxUnavailable
    const parseStrategyValue = (value, total) => {
        if (!value) return 0;
        if (typeof value === 'number') return value;
        if (typeof value === 'string') {
            if (value.endsWith('%')) {
                const percent = parseFloat(value);
                return Math.ceil((total * percent) / 100);
            }
            return parseInt(value) || 0;
        }
        return 0;
    };

    const maxSurge = strategy ? parseStrategyValue(strategy.maxSurge, totalReplicas) : Math.ceil(totalReplicas * 0.25);
    const maxUnavailable = strategy ? parseStrategyValue(strategy.maxUnavailable, totalReplicas) : Math.floor(totalReplicas * 0.25);
    const strategyType = strategy?.type || 'RollingUpdate';

    // Determine behavior message based on strategy and replica count
    let behaviorMessage = '';
    let behaviorColor = 'text-yellow-400';
    let strategyDetails = '';

    if (totalReplicas === 0) {
        behaviorMessage = 'Warning: This deployment has 0 replicas. Rollout will have no effect.';
        behaviorColor = 'text-gray-400';
    } else if (strategyType === 'Recreate') {
        behaviorMessage = `This will use the Recreate strategy: all existing pods will be terminated before new pods are created.`;
        behaviorColor = 'text-orange-400';
        strategyDetails = `⚠️ Service will be unavailable during the rollout (approximately ${totalReplicas} pod(s) will be restarted sequentially).`;
    } else if (strategyType === 'RollingUpdate') {
        // Calculate actual pods that can be unavailable
        const minAvailable = Math.max(0, totalReplicas - maxUnavailable);
        const maxTotal = totalReplicas + maxSurge;

        if (totalReplicas === 1) {
            behaviorMessage = 'This will restart the single pod, causing a brief service interruption.';
            behaviorColor = 'text-orange-400';
            strategyDetails = `⚠️ Service will be unavailable during the restart (1 pod will be replaced).`;
        } else {
            behaviorMessage = `This will use the RollingUpdate strategy: pods will be updated gradually while maintaining service availability.`;
            behaviorColor = 'text-green-400';
            strategyDetails = `✓ Kubernetes will maintain at least ${minAvailable} pod(s) available during the update.\n✓ Up to ${maxSurge} new pod(s) can be created above the desired count.\n✓ Up to ${maxUnavailable} pod(s) can be unavailable during the update.\n✓ Maximum ${maxTotal} pod(s) can exist simultaneously during the rollout.`;
        }
    } else {
        behaviorMessage = `This will restart ${totalReplicas} pods using the ${strategyType} strategy.`;
        behaviorColor = 'text-yellow-400';
    }

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl max-h-[90vh] overflow-y-auto">
                <h3 className="text-lg font-semibold text-white mb-2">
                    Confirm rollout
                </h3>
                <p className="text-sm text-gray-300 mb-3">
                    Rollout {confirmRollout.kind} "<span className="font-medium text-white">{confirmRollout.name}</span>"?
                </p>
                {details && (
                    <div className="mb-3 p-3 bg-gray-800/50 rounded border border-gray-700">
                        <div className="text-xs text-gray-400 mb-2">Deployment Information:</div>
                        <div className="text-sm text-gray-300 space-y-1">
                            <div className="flex justify-between">
                                <span className="text-gray-400">Total Replicas:</span>
                                <span className="font-medium text-white">{totalReplicas}</span>
                            </div>
                            <div className="flex justify-between">
                                <span className="text-gray-400">Ready Pods:</span>
                                <span className={`font-medium ${readyCount === totalReplicas ? 'text-green-400' : 'text-yellow-400'}`}>
                                    {readyCount} / {totalReplicas}
                                </span>
                            </div>
                            {strategy && (
                                <div className="flex justify-between">
                                    <span className="text-gray-400">Update Strategy:</span>
                                    <span className="font-medium text-blue-400">{strategyType}</span>
                                </div>
                            )}
                            {strategy && strategyType === 'RollingUpdate' && (
                                <>
                                    <div className="flex justify-between">
                                        <span className="text-gray-400">Max Surge:</span>
                                        <span className="font-medium text-white">{strategy.maxSurge || '25%'}</span>
                                    </div>
                                    <div className="flex justify-between">
                                        <span className="text-gray-400">Max Unavailable:</span>
                                        <span className="font-medium text-white">{strategy.maxUnavailable || '25%'}</span>
                                    </div>
                                </>
                            )}
                        </div>
                    </div>
                )}
                <div className={`text-xs mb-3 p-3 rounded border ${behaviorColor.includes('green') ? 'bg-green-900/20 border-green-700/50' : behaviorColor.includes('orange') ? 'bg-orange-900/20 border-orange-700/50' : 'bg-yellow-900/20 border-yellow-700/50'}`}>
                    <p className={`font-medium mb-1 ${behaviorColor}`}>
                        {behaviorMessage}
                    </p>
                    {strategyDetails && (
                        <div className="mt-2 text-xs text-gray-300 whitespace-pre-line">
                            {strategyDetails}
                        </div>
                    )}
                </div>
                <div className="flex justify-end space-x-3">
                    <button
                        onClick={() => setConfirmRollout(null)}
                        className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                    >
                        Cancel
                    </button>
                    <button
                        onClick={() => handleRolloutDeployment(confirmRollout)}
                        disabled={totalReplicas === 0}
                        className="px-4 py-2 bg-green-600 hover:bg-green-700 text-white rounded-md transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                    >
                        Rollout
                    </button>
                </div>
            </div>
        </div>
    );
};

export default RolloutConfirmationModal;
