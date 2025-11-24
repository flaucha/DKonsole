import { useQuery } from '@tanstack/react-query';
import { fetchWorkloads } from '../api/k8sApi';
import { useSettings } from '../context/SettingsContext';

export const useWorkloads = (fetcher, namespace, kind, continueToken = null) => {
    const { itemsPerPage } = useSettings();
    const isEnabled = !!namespace && !!kind;
    return useQuery({
        queryKey: ['workloads', namespace, kind, itemsPerPage, continueToken],
        queryFn: () => fetchWorkloads(fetcher, namespace, kind, itemsPerPage, continueToken),
        enabled: isEnabled,
        refetchInterval: isEnabled ? 5000 : false, // Poll every 5 seconds only when enabled
    });
};
