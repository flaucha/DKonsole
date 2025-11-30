import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { fetchHelmReleases } from '../api/k8sApi';

export const useHelmReleases = (fetcher, currentCluster) => {
    return useQuery({
        queryKey: ['helmReleases', currentCluster],
        queryFn: () => fetchHelmReleases(fetcher, currentCluster),
        refetchInterval: 10000, // Reduced to 10 seconds
        placeholderData: keepPreviousData,
    });
};
