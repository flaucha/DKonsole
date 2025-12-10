import React, { useMemo, useState } from 'react';
import { useColumnOrder } from '../../hooks/useColumnOrder';
import { useNamespaceColumns } from '../../hooks/useNamespaceColumns';
import NamespaceTable from './NamespaceTable';

const NamespaceList = ({
    namespaces,
    sortField,
    setSortField,
    sortDirection,
    setSortDirection,
    expandedId,
    setExpandedId,
    isAdmin,
    onEditYaml,
    onDelete,
    user
}) => {
    const [menuOpen, setMenuOpen] = useState(null);

    const toggleExpand = (nsName) => {
        setExpandedId(current => current === nsName ? null : nsName);
    };

    const handleSort = (field) => {
        setSortField((prevField) => {
            if (prevField === field) {
                setSortDirection((prevDir) => (prevDir === 'asc' ? 'desc' : 'asc'));
                return prevField;
            }
            setSortDirection('asc');
            return field;
        });
    };

    const { dataColumns, ageColumn, actionsColumn } = useNamespaceColumns({
        isAdmin,
        onEditYaml,
        onDelete,
        menuOpen,
        setMenuOpen
    });

    const reorderableColumns = useMemo(
        () => dataColumns.filter((col) => !col.pinned && !col.isAction),
        [dataColumns]
    );

    const { orderedColumns, moveColumn } = useColumnOrder(reorderableColumns, 'namespace-columns', user?.username);

    const sortableColumns = useMemo(
        () => [...dataColumns, ageColumn].filter((col) => typeof col.sortValue === 'function'),
        [dataColumns, ageColumn]
    );

    const sortedNamespaces = useMemo(() => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const activeColumn = sortableColumns.find((col) => col.id === sortField) || sortableColumns[0];
        if (!activeColumn) return namespaces;
        return [...namespaces].sort((a, b) => {
            const va = activeColumn.sortValue(a);
            const vb = activeColumn.sortValue(b);
            if (typeof va === 'number' && typeof vb === 'number') {
                return (va - vb) * dir;
            }
            return String(va).localeCompare(String(vb)) * dir;
        });
    }, [namespaces, sortDirection, sortField, sortableColumns]);

    const columns = useMemo(
        () => [...orderedColumns, ageColumn, actionsColumn],
        [orderedColumns, ageColumn, actionsColumn]
    );

    const gridTemplateColumns = useMemo(
        () => columns.map((col) => col.width || 'minmax(120px, 1fr)').join(' '),
        [columns]
    );

    return (
        <NamespaceTable
            namespaces={sortedNamespaces}
            columns={columns}
            gridTemplateColumns={gridTemplateColumns}
            expandedId={expandedId}
            toggleExpand={toggleExpand}
            sortField={sortField}
            sortDirection={sortDirection}
            handleSort={handleSort}
            moveColumn={moveColumn}
            onEditYaml={onEditYaml}
            menuOpen={menuOpen}
            setMenuOpen={setMenuOpen}
        />
    );
};

export default NamespaceList;

