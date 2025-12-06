import React from 'react';
import { Activity, HardDrive } from 'lucide-react';
import ResourceMetrics from './shared/ResourceMetrics';

const DeploymentMetrics = ({ deployment, namespace }) => {
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
        }
    ];

    return (
        <ResourceMetrics
            resourceName={deployment?.name}
            namespace={namespace}
            resourceType="deployment"
            apiEndpoint="/api/prometheus/metrics"
            chartConfigs={chartConfigs}
        />
    );
};

export default DeploymentMetrics;
