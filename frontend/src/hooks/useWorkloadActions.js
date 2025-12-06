import { useState } from 'react';

export const useWorkloadActions = (authFetch, refetch, currentCluster) => {
    const [confirmAction, setConfirmAction] = useState(null);
    const [confirmRollout, setConfirmRollout] = useState(null);
    const [scaling, setScaling] = useState(null);
    const [triggering, setTriggering] = useState(null);
    const [createdJob, setCreatedJob] = useState(null);
    const [rollingOut, setRollingOut] = useState(null);

    const handleDelete = async (res, force = false) => {
        const params = new URLSearchParams({ kind: res.kind, name: res.name });
        if (res.namespace) params.append('namespace', res.namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        if (force) params.append('force', 'true');

        try {
            const response = await authFetch(`/api/resource?${params.toString()}`, { method: 'DELETE' });
            if (!response.ok) {
                const errorText = await response.text().catch(() => 'Delete failed');
                throw new Error(errorText || 'Delete failed');
            }
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to delete resource');
        } finally {
            setConfirmAction(null);
        }
    };

    const handleScale = async (res, delta) => {
        if (!res.namespace) return;
        setScaling(res.name);
        const params = new URLSearchParams({
            kind: 'Deployment',
            name: res.name,
            namespace: res.namespace,
            delta: String(delta),
        });
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/scale?${params.toString()}`, { method: 'POST' });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Scale failed');
                throw new Error(errorText || 'Scale failed');
            }
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to scale deployment');
        } finally {
            setScaling(null);
        }
    };

    const handleTriggerCronJob = async (res) => {
        if (!res.namespace) return;
        setTriggering(res.name);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/cronjobs/trigger?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    namespace: res.namespace,
                    name: res.name
                })
            });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Trigger failed');
                throw new Error(errorText || 'Trigger failed');
            }
            const data = await resp.json();
            if (data.jobName) {
                setCreatedJob({
                    name: data.jobName,
                    namespace: res.namespace
                });
            }
        } catch (err) {
            alert(err.message || 'Failed to trigger cronjob');
        } finally {
            setTriggering(null);
        }
    };

    const handleRolloutDeployment = async (res) => {
        if (!res.namespace) return;
        setRollingOut(res.name);
        setConfirmRollout(null);
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        try {
            const resp = await authFetch(`/api/deployments/rollout?${params.toString()}`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    namespace: res.namespace,
                    name: res.name
                })
            });
            if (!resp.ok) {
                const errorText = await resp.text().catch(() => 'Rollout failed');
                throw new Error(errorText || 'Rollout failed');
            }
            // Refresh the list to show updated deployment
            refetch();
        } catch (err) {
            alert(err.message || 'Failed to rollout deployment');
        } finally {
            setRollingOut(null);
        }
    };

    return {
        confirmAction,
        setConfirmAction,
        confirmRollout,
        setConfirmRollout,
        scaling,
        triggering,
        createdJob,
        setCreatedJob,
        rollingOut,
        handleDelete,
        handleScale,
        handleTriggerCronJob,
        handleRolloutDeployment
    };
};
