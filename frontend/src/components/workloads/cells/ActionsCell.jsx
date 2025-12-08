import React, { useState } from 'react';
import { PlayCircle, RefreshCw, MoreVertical, Edit2 } from 'lucide-react';
import { isAdmin, canEdit } from '../../../utils/permissions';
import { isMasterNode } from '../../../utils/workloadUtils';

const ActionsCell = ({
    res,
    kind,
    user,
    currentCluster,
    authFetch,
    handleTriggerCronJob,
    triggering,
    rollingOut,
    setConfirmAction,
    setConfirmRollout,
    onEditYaml,
    onEditInPlace
}) => {
    const [menuOpen, setMenuOpen] = useState(false);

    // This logic handles the "Rollout" button click which fetches deployment details
    // before showing the confirmation modal.
    const handleRolloutClick = async (e) => {
        e.stopPropagation();
        try {
            const params = new URLSearchParams();
            if (currentCluster) params.append('cluster', currentCluster);
            const response = await authFetch(
                `/ api / namespaces / ${res.namespace} /Deployment/${res.name}?${params.toString()} `
            );
            if (response.ok) {
                const deploymentData = await response.json();
                let strategyInfo = null;
                if (deploymentData.raw && deploymentData.raw.spec) {
                    const spec = deploymentData.raw.spec;
                    const strategy = spec.strategy || {};
                    const strategyType = strategy.type || 'RollingUpdate';

                    if (strategyType === 'RollingUpdate') {
                        const rollingUpdate = strategy.rollingUpdate || {};
                        strategyInfo = {
                            type: 'RollingUpdate',
                            maxSurge: rollingUpdate.maxSurge || '25%',
                            maxUnavailable: rollingUpdate.maxUnavailable || '25%'
                        };
                    } else {
                        strategyInfo = {
                            type: 'Recreate'
                        };
                    }
                }
                setConfirmRollout({
                    ...res,
                    details: deploymentData.details || res.details,
                    strategy: strategyInfo
                });
            } else {
                setConfirmRollout(res);
            }
        } catch {
            setConfirmRollout(res);
        }
    };

    return (
        <div className="flex justify-end items-center space-x-1 pr-2 flex-nowrap shrink-0 min-w-0" onClick={(e) => e.stopPropagation()}>
            {kind === 'CronJob' && (
                <button
                    onClick={(e) => {
                        e.stopPropagation();
                        handleTriggerCronJob(res);
                    }}
                    disabled={triggering === res.name}
                    className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-green-400 transition-colors disabled:opacity-50"
                    title="Trigger manual run"
                >
                    <PlayCircle size={16} />
                </button>
            )}
            {kind === 'Deployment' && (isAdmin(user) || canEdit(user, res.namespace)) && (
                <button
                    onClick={handleRolloutClick}
                    disabled={rollingOut === res.name}
                    className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-green-400 transition-colors disabled:opacity-50"
                    title="Rollout deployment"
                >
                    <RefreshCw size={16} className={rollingOut === res.name ? 'animate-spin' : ''} />
                </button>
            )}
            <div className="relative">
                <button
                    onClick={() => setMenuOpen(!menuOpen)}
                    className="p-1 hover:bg-gray-800 rounded text-gray-400 hover:text-white transition-colors"
                >
                    <MoreVertical size={16} />
                </button>
                {menuOpen && (
                    <>
                        <div
                            className="fixed inset-0 z-40"
                            onClick={() => setMenuOpen(false)}
                        />
                        <div className="absolute right-0 mt-1 w-48 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50 py-1">
                            <div className="flex flex-col">
                                {(isAdmin(user) || canEdit(user, res.namespace)) && (kind === 'ConfigMap' || kind === 'Secret') && onEditInPlace && (
                                    <button
                                        onClick={() => {
                                            setMenuOpen(false);
                                            onEditInPlace(res);
                                        }}
                                        className="flex items-center px-4 py-2 text-sm text-gray-200 hover:bg-gray-700 w-full text-left"
                                    >
                                        <Edit2 size={14} className="mr-2" />
                                        Edit in place
                                    </button>
                                )}
                                {isAdmin(user) || canEdit(user, res.namespace) ? (
                                    <>
                                        {/* Unified Actions */}
                                        {onEditYaml && (
                                            <button
                                                onClick={() => {
                                                    onEditYaml(res);
                                                    setMenuOpen(false);
                                                }}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700 flex items-center gap-2"
                                            >
                                                Edit YAML
                                            </button>
                                        )}

                                        {kind === 'Deployment' && (
                                            <button
                                                onClick={(e) => {
                                                    handleRolloutClick(e);
                                                    setMenuOpen(false);
                                                }}
                                                disabled={rollingOut === res.name}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700 flex items-center gap-2"
                                            >
                                                Restart Rollout
                                            </button>
                                        )}

                                        {kind === 'CronJob' && (
                                            <button
                                                onClick={(e) => {
                                                    e.stopPropagation();
                                                    handleTriggerCronJob(res);
                                                    setMenuOpen(false);
                                                }}
                                                disabled={triggering === res.name}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700 flex items-center gap-2"
                                            >
                                                Trigger Run
                                            </button>
                                        )}

                                        <div className="h-px bg-gray-700 my-1 mx-2" />

                                        {res.kind !== 'Node' || (res.kind === 'Node' && !isMasterNode(res)) ? (
                                            <button
                                                onClick={() => {
                                                    setConfirmAction({ res, force: false });
                                                    setMenuOpen(false);
                                                }}
                                                className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                            >
                                                Delete
                                            </button>
                                        ) : null}

                                        {res.kind !== 'Node' && (
                                            <button
                                                onClick={() => {
                                                    setConfirmAction({ res, force: true });
                                                    setMenuOpen(false);
                                                }}
                                                className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                            >
                                                Force Delete
                                            </button>
                                        )}
                                    </>
                                ) : (
                                    <div className="px-4 py-2 text-xs text-gray-500">
                                        View only
                                    </div>
                                )}
                            </div>
                        </div>
                    </>
                )}
            </div>
        </div>
    );
};

export default ActionsCell;
