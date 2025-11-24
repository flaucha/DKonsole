import { useQuery } from '@tanstack/react-query';
import { fetchHelmReleases } from '../api/k8sApi';

export const useHelmReleases = (fetcher, currentCluster) => {
    return useQuery({
        queryKey: ['helmReleases', currentCluster],
        queryFn: () => fetchHelmReleases(fetcher, currentCluster),
        refetchInterval: 30000, // 30 seconds
    });
};
