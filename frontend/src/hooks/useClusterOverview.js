import { useQuery } from '@tanstack/react-query';
import { fetchClusterOverview, fetchPrometheusStatus, fetchClusterMetrics } from '../api/k8sApi';

export const useClusterOverview = (fetcher, currentCluster) => {
    const overviewQuery = useQuery({
        queryKey: ['clusterOverview'],
        queryFn: () => fetchClusterOverview(fetcher),
        refetchInterval: 30000, // 30 seconds
    });

    const prometheusStatusQuery = useQuery({
        queryKey: ['prometheusStatus', currentCluster],
        queryFn: () => fetchPrometheusStatus(fetcher, currentCluster),
        enabled: !!currentCluster,
    });

    const prometheusEnabled = prometheusStatusQuery.data?.enabled || false;

    const metricsQuery = useQuery({
        queryKey: ['clusterMetrics', currentCluster],
        queryFn: () => fetchClusterMetrics(fetcher, currentCluster),
        enabled: !!currentCluster && prometheusEnabled,
        refetchInterval: 15000, // 15 seconds
    });

    return {
        overview: overviewQuery,
        prometheusStatus: prometheusStatusQuery,
        metrics: metricsQuery,
    };
};
