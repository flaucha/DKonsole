import { useQuery } from '@tanstack/react-query';
import { fetchNamespaces } from '../api/k8sApi';

export const useNamespaces = (fetcher, currentCluster) => {
    return useQuery({
        queryKey: ['namespaces', currentCluster],
        queryFn: () => fetchNamespaces(fetcher, currentCluster),
        refetchInterval: 30000, // 30 seconds
    });
};
