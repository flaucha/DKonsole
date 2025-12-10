import {
    Box,
    Server,
    Shield,
    Lock,
    Users,
    FileText,
    Key,
    Clock,
    Layers,
    Activity,
    Network,
    Globe,
    HardDrive
} from 'lucide-react';


// Map resource kind to an icon component
export const getIcon = (kind) => {
    switch (kind) {
        case 'Deployment':
        case 'Pod':
            return Box;
        case 'Node':
            return Server;
        case 'ServiceAccount':
            return Shield;
        case 'Role':
        case 'ClusterRole':
            return Lock;
        case 'RoleBinding':
        case 'ClusterRoleBinding':
            return Users;
        case 'ConfigMap':
            return FileText;
        case 'Secret':
            return Key;
        case 'Job':
        case 'CronJob':
            return Clock;
        case 'StatefulSet':
        case 'DaemonSet':
            return Layers;
        case 'HPA':
            return Activity;
        case 'Service':
            return Network;
        case 'Ingress':
            return Globe;
        case 'NetworkPolicy':
            return Activity;
        case 'PersistentVolumeClaim':
        case 'PersistentVolume':
            return HardDrive;
        case 'StorageClass':
            return HardDrive;
        default:
            return Box;
    }
};

export const parseDateValue = (value) => {
    const timestamp = new Date(value).getTime();
    return Number.isFinite(timestamp) ? timestamp : 0;
};

export const parseCpuToMilli = (cpuStr) => {
    if (!cpuStr) return 0;
    const trimmed = cpuStr.trim();
    if (trimmed.endsWith('m')) return parseFloat(trimmed.replace('m', '')) || 0;
    const val = parseFloat(trimmed);
    return Number.isFinite(val) ? val * 1000 : 0;
};

export const parseMemoryToMi = (memStr) => {
    if (!memStr) return 0;
    const normalized = memStr.toUpperCase().trim();
    const num = parseFloat(normalized);
    if (!Number.isFinite(num)) return 0;
    if (normalized.includes('TI') || normalized.includes('T')) return num * 1024 * 1024;
    if (normalized.includes('GI') || normalized.includes('G')) return num * 1024;
    if (normalized.includes('MI') || normalized.includes('M')) return num;
    if (normalized.includes('KI') || normalized.includes('K')) return num / 1024;
    return num;
};

export const parseSizeToMi = (sizeStr) => {
    if (!sizeStr) return 0;
    const normalized = sizeStr.toUpperCase().trim();
    if (!normalized || normalized === '—') return 0;
    const num = parseFloat(normalized);
    if (!Number.isFinite(num)) return 0;
    if (normalized.includes('TI') || normalized.includes('T')) return num * 1024 * 1024;
    if (normalized.includes('GI') || normalized.includes('G')) return num * 1024;
    if (normalized.includes('MI') || normalized.includes('M')) return num;
    if (normalized.includes('KI') || normalized.includes('K')) return num / 1024;
    return num / (1024 * 1024);
};

export const parseReadyRatio = (readyStr) => {
    if (!readyStr) return 0;
    const parts = readyStr.toString().split('/');
    if (parts.length !== 2) return 0;
    const ready = parseFloat(parts[0]) || 0;
    const total = parseFloat(parts[1]) || 0;
    return total > 0 ? ready / total : 0;
};

// Helper function to check if a node is a master/control-plane node
export const isMasterNode = (res) => {
    if (res.kind !== 'Node') return false;

    // Check details for labels or taints that indicate control plane
    if (res.details) {
        // Check labels if available
        if (res.details.labels) {
            const labels = res.details.labels;
            // distinct check for existence (key present) vs value
            const cpKey = 'node-role.kubernetes.io/control-plane';
            const masterKey = 'node-role.kubernetes.io/master';

            if (Object.prototype.hasOwnProperty.call(labels, cpKey)) {
                if (labels[cpKey] !== 'false') return true;
            }
            if (Object.prototype.hasOwnProperty.call(labels, masterKey)) {
                if (labels[masterKey] !== 'false') return true;
            }
        }

        // Check taints if available
        if (res.details.taints && Array.isArray(res.details.taints)) {
            for (const taint of res.details.taints) {
                const taintKey = taint.key || (typeof taint === 'string' ? taint : null);
                if (taintKey && (
                    taintKey === 'node-role.kubernetes.io/control-plane' ||
                    taintKey === 'node-role.kubernetes.io/master')) {
                    return true;
                }
            }
        }
    }

    return false;
};
