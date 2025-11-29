import { useQuery, keepPreviousData } from '@tanstack/react-query';
import { fetchClusterOverview, fetchPrometheusStatus, fetchClusterMetrics } from '../api/k8sApi';

export const useClusterOverview = (fetcher, currentCluster) => {
    const overviewQuery = useQuery({
        queryKey: ['clusterOverview'],
        queryFn: () => fetchClusterOverview(fetcher),
        refetchInterval: 10000, // Reduced to 10 seconds
        placeholderData: keepPreviousData,
    });

    const prometheusStatusQuery = useQuery({
        queryKey: ['prometheusStatus', currentCluster],
        queryFn: () => fetchPrometheusStatus(fetcher, currentCluster),
        enabled: !!currentCluster,
        placeholderData: keepPreviousData,
    });

    const prometheusEnabled = prometheusStatusQuery.data?.enabled || false;

    const metricsQuery = useQuery({
        queryKey: ['clusterMetrics', currentCluster],
        queryFn: () => fetchClusterMetrics(fetcher, currentCluster),
        enabled: !!currentCluster && prometheusEnabled,
        refetchInterval: 5000, // Reduced to 5 seconds for metrics
        placeholderData: keepPreviousData,
    });

    return {
        overview: overviewQuery,
        prometheusStatus: prometheusStatusQuery,
        metrics: metricsQuery,
    };
};
