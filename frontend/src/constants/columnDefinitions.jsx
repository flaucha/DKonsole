import { getBaseColumns, getAgeColumn } from './columns/common';
import { podColumns } from './columns/pod';
import { deploymentColumns } from './columns/deployment';
import { serviceColumns } from './columns/service';
import { ingressColumns } from './columns/ingress';
import { nodeColumns } from './columns/node';
import { jobColumns } from './columns/job';
import { cronJobColumns } from './columns/cronJob';
import {
    persistentVolumeColumns,
    persistentVolumeClaimColumns,
    storageClassColumns
} from './columns/storage';
import { configColumns } from './columns/config';
import { networkPolicyColumns } from './columns/networkPolicy';

export const getWorkloadColumns = (kind) => {
    // Common columns
    const baseColumns = getBaseColumns(kind);

    // Kind-specific columns
    let specificColumns = [];

    switch (kind) {
        case 'Pod':
            specificColumns = podColumns;
            break;
        case 'Deployment':
            specificColumns = deploymentColumns;
            break;
        case 'Service':
            specificColumns = serviceColumns;
            break;
        case 'Ingress':
            specificColumns = ingressColumns;
            break;
        case 'Node':
            specificColumns = nodeColumns;
            break;
        case 'Job':
            specificColumns = jobColumns;
            break;
        case 'CronJob':
            specificColumns = cronJobColumns;
            break;
        case 'PersistentVolumeClaim':
            specificColumns = persistentVolumeClaimColumns;
            break;
        case 'PersistentVolume':
            specificColumns = persistentVolumeColumns;
            break;
        case 'StorageClass':
            specificColumns = storageClassColumns;
            break;
        case 'ConfigMap':
        case 'Secret':
            specificColumns = configColumns;
            break;
        case 'NetworkPolicy':
            specificColumns = networkPolicyColumns;
            break;
        default:
            specificColumns = [];
            break;
    }

    const ageColumn = getAgeColumn();

    return [...baseColumns, ...specificColumns, ageColumn];
};
