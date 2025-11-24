export const fetchWorkloads = async (fetcher, namespace, kind) => {
    if (!kind) {
        throw new Error('Resource kind is required');
    }
    if (!namespace) {
        throw new Error('Namespace is required');
    }
    const f = fetcher || fetch;
    // Use /api/resources endpoint which expects kind as query parameter
    const params = new URLSearchParams({ kind });
    params.append('namespace', namespace);
    const response = await f(`/api/resources?${params.toString()}`);
    if (!response.ok) {
        const errorText = await response.text().catch(() => 'Unknown error');
        throw new Error(`Failed to fetch ${kind}: ${errorText}`);
    }
    return response.json();
};

export const fetchPodLogs = async (fetcher, namespace, pod, container) => {
    const f = fetcher || fetch;
    const response = await f(`/api/pods/logs?namespace=${namespace}&pod=${pod}&container=${container || ''}`);
    if (!response.ok) {
        throw new Error('Failed to fetch logs');
    }
    return response.body; // Returns a ReadableStream
};

export const fetchResource = async (fetcher, kind, name, namespace) => {
    const f = fetcher || fetch;
    const url = namespace
        ? `/api/namespaces/${namespace}/${kind}/${name}`
        : `/api/cluster/${kind}/${name}`;

    const response = await f(url);
    if (!response.ok) {
        throw new Error(`Failed to fetch ${kind} ${name}`);
    }
    return response.json();
};

export const fetchResources = async (fetcher, kind, namespace, cluster) => {
    const f = fetcher || fetch;
    const params = new URLSearchParams({ kind });
    if (namespace && namespace !== 'all') params.append('namespace', namespace);
    if (cluster) params.append('cluster', cluster);

    const response = await f(`/api/resources?${params.toString()}`);
    if (!response.ok) {
        throw new Error(`Failed to fetch ${kind}`);
    }
    return response.json();
};

export const fetchClusterOverview = async (fetcher) => {
    const f = fetcher || fetch;
    const response = await f('/api/overview');
    if (!response.ok) {
        throw new Error('Failed to fetch cluster overview');
    }
    return response.json();
};

export const fetchPrometheusStatus = async (fetcher, cluster) => {
    const f = fetcher || fetch;
    const params = new URLSearchParams();
    if (cluster) params.append('cluster', cluster);

    const response = await f(`/api/prometheus/status?${params.toString()}`);
    if (!response.ok) {
        throw new Error('Failed to fetch Prometheus status');
    }
    return response.json();
};

export const fetchClusterMetrics = async (fetcher, cluster) => {
    const f = fetcher || fetch;
    const params = new URLSearchParams();
    if (cluster) params.append('cluster', cluster);

    const response = await f(`/api/prometheus/cluster-overview?${params.toString()}`);
    if (!response.ok) {
        throw new Error('Failed to fetch cluster metrics');
    }
    return response.json();
};

export const fetchNamespaces = async (fetcher, cluster) => {
    const f = fetcher || fetch;
    const params = new URLSearchParams();
    if (cluster) params.append('cluster', cluster);

    const response = await f(`/api/namespaces?${params.toString()}`);
    if (!response.ok) {
        throw new Error('Failed to fetch namespaces');
    }
    return response.json();
};

export const fetchHelmReleases = async (fetcher, cluster) => {
    const f = fetcher || fetch;
    const params = new URLSearchParams();
    if (cluster) params.append('cluster', cluster);

    const response = await f(`/api/helm/releases?${params.toString()}`);
    if (!response.ok) {
        throw new Error('Failed to fetch helm releases');
    }
    return response.json();
};
