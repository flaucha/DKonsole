import { useState, useEffect } from 'react';
import { parseErrorResponse, parseError } from '../utils/errorParser';
import { getResourceTemplate } from '../utils/yamlTemplates';

export const useYamlEditor = (resource, currentCluster, authFetch, onSaved) => {
    const { name, namespace, kind, group, version, resource: resourceName, namespaced } = resource || {};
    const [content, setContent] = useState('');
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [error, setError] = useState('');

    const buildUrl = () => {
        const params = new URLSearchParams({ kind, name });
        if (namespace) params.append('namespace', namespace);
        if (group) params.append('group', group);
        if (version) params.append('version', version);
        if (resourceName) params.append('resource', resourceName);
        if (namespaced !== undefined) params.append('namespaced', namespaced.toString());
        if (currentCluster) params.append('cluster', currentCluster);
        return `/api/resource/yaml?${params.toString()}`;
    };

    useEffect(() => {
        if (!resource) return;

        // If it's a new resource, provide a template with proper spec
        if (resource.isNew) {
            const ns = namespace || resource.metadata?.namespace || 'dkonsole';
            const template = getResourceTemplate(kind, ns);
            setContent(template);
            setLoading(false);
            return;
        }

        setLoading(true);
        setError('');
        authFetch(buildUrl())
            .then(async (res) => {
                if (!res.ok) {
                    const text = await parseErrorResponse(res);
                    throw new Error(text || 'Failed to load resource');
                }
                return res.text();
            })
            .then((yaml) => {
                setContent(yaml);
                setLoading(false);
            })
            .catch((err) => {
                setError(parseError(err));
                setLoading(false);
            });
    }, [resource, currentCluster]);

    const handleSave = () => {
        setSaving(true);
        setError('');

        // Use import endpoint which uses Server-Side Apply (equivalent to kubectl apply -f)
        const params = new URLSearchParams();
        if (currentCluster) params.append('cluster', currentCluster);

        authFetch(`/api/resource/import?${params.toString()}`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/x-yaml' },
            body: content,
        })
            .then(async (res) => {
                if (!res.ok) {
                    const text = await parseErrorResponse(res);
                    throw new Error(text || 'Failed to apply resource');
                }
                setSaving(false);
                onSaved?.();
            })
            .catch((err) => {
                setError(parseError(err));
                setSaving(false);
            });
    };

    return {
        content,
        setContent,
        loading,
        saving,
        error,
        setError,
        handleSave
    };
};
