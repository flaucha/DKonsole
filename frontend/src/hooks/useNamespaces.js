import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { fetchNamespaces } from '../api/k8sApi';

export const useNamespaces = (fetcher, currentCluster) => {
    return useQuery({
        queryKey: ['namespaces', currentCluster],
        queryFn: () => fetchNamespaces(fetcher, currentCluster),
        refetchInterval: 10000, // Reduced to 10 seconds
        placeholderData: keepPreviousData,
    });
};
