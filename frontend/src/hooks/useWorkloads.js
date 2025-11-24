import { useQuery } from '@tanstack/react-query';
import { fetchWorkloads } from '../api/k8sApi';

export const useWorkloads = (fetcher, namespace, kind) => {
    const isEnabled = !!namespace && !!kind;
    return useQuery({
        queryKey: ['workloads', namespace, kind],
        queryFn: () => fetchWorkloads(fetcher, namespace, kind),
        enabled: isEnabled,
        refetchInterval: isEnabled ? 5000 : false, // Poll every 5 seconds only when enabled
    });
};
