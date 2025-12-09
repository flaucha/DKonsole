import { useMemo } from 'react';
import { getWorkloadColumns } from '../constants/columnDefinitions';

const useWorkloadColumns = (kind) => {
    // Memoize the dataColumns to prevent unnecessary re-renders
    const dataColumns = useMemo(() => {
        return getWorkloadColumns(kind);
    }, [kind]);

    return { dataColumns };
};

export default useWorkloadColumns;
