import React from 'react';
import { Activity, HardDrive, Network, Database } from 'lucide-react';
import ResourceMetrics from './shared/ResourceMetrics';

const PodMetrics = ({ pod, namespace }) => {
    const chartConfigs = [
        {
            title: 'CPU (millicores)',
            dataKey: 'cpu',
            color: '#60A5FA',
            icon: Activity,
            alwaysVisible: true
        },
        {
            title: 'Memory (MiB)',
            dataKey: 'memory',
            color: '#A78BFA',
            icon: HardDrive,
            alwaysVisible: true
        },
        {
            title: 'Network RX (KB/s)',
            dataKey: 'networkRx',
            color: '#34D399',
            icon: Network,
            alwaysVisible: false
        },
        {
            title: 'Network TX (KB/s)',
            dataKey: 'networkTx',
            color: '#FBBF24',
            icon: Network,
            alwaysVisible: false
        },
        {
            title: 'PVC Usage (%)',
            dataKey: 'pvcUsage',
            color: '#FB923C',
            icon: Database,
            unit: '%',
            alwaysVisible: false
        }
    ];

    return (
        <ResourceMetrics
            resourceName={pod?.name}
            namespace={namespace}
            resourceType="pod"
            apiEndpoint="/api/prometheus/pod-metrics"
            chartConfigs={chartConfigs}
        />
    );
};

export default PodMetrics;

