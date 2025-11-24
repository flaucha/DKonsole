import { useQuery } from '@tanstack/react-query';
import { fetchWorkloads } from '../api/k8sApi';

export const useWorkloads = (namespace, kind) => {
    return useQuery({
        queryKey: ['workloads', namespace, kind],
        queryFn: () => fetchWorkloads(namespace, kind),
        enabled: !!namespace && !!kind,
        refetchInterval: 5000, // Poll every 5 seconds
    });
};
