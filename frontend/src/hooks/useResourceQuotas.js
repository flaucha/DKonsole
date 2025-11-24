import { useQuery } from '@tanstack/react-query';
import { fetchResources } from '../api/k8sApi';

export const useResourceQuotas = (fetcher, namespace, currentCluster) => {
    const quotasQuery = useQuery({
        queryKey: ['resourceQuotas', namespace, currentCluster],
        queryFn: () => fetchResources(fetcher, 'ResourceQuota', namespace, currentCluster),
        refetchInterval: 10000,
    });

    const limitRangesQuery = useQuery({
        queryKey: ['limitRanges', namespace, currentCluster],
        queryFn: () => fetchResources(fetcher, 'LimitRange', namespace, currentCluster),
        refetchInterval: 10000,
    });

    return {
        quotas: quotasQuery,
        limitRanges: limitRangesQuery,
    };
};
