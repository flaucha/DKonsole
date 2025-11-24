import { useQuery } from '@tanstack/react-query';
import { fetchWorkloads } from '../api/k8sApi';

export const useWorkloads = (fetcher, namespace, kind, currentCluster) => {
    const isEnabled = !!namespace && !!kind;
    return useQuery({
        queryKey: ['workloads', namespace, kind, currentCluster],
        queryFn: () => fetchWorkloads(fetcher, namespace, kind, currentCluster),
        enabled: isEnabled,
        refetchInterval: isEnabled ? 5000 : false, // Poll every 5 seconds only when enabled
        staleTime: 0, // Data is immediately stale, forcing fresh fetch on mount
        refetchOnMount: true, // Always refetch when component mounts
        refetchOnWindowFocus: false, // Don't refetch on window focus to avoid unnecessary requests
    });
};
