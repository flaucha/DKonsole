import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { fetchWorkloads } from '../api/k8sApi';

export const useWorkloads = (fetcher, namespace, kind, currentCluster) => {
    const isEnabled = !!namespace && !!kind;
    return useQuery({
        queryKey: ['workloads', namespace, kind, currentCluster],
        queryFn: () => fetchWorkloads(fetcher, namespace, kind, currentCluster),
        enabled: isEnabled,
        refetchInterval: isEnabled ? 2000 : false, // Poll every 2 seconds for snappier updates
        staleTime: 1000, // Data is fresh for 1 second
        placeholderData: keepPreviousData, // Keep previous data while fetching new data
        refetchOnMount: true,
        refetchOnWindowFocus: false,
    });
};
