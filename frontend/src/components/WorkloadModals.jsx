import React from 'react';
import YamlEditor from '../components/YamlEditor';
import { DataEditor } from '../components/details/DataEditor';
import DeleteConfirmationModal from '../components/workloads/modals/DeleteConfirmationModal';
import RolloutConfirmationModal from '../components/workloads/modals/RolloutConfirmationModal';
import JobCreatedModal from '../components/workloads/modals/JobCreatedModal';

const WorkloadModals = ({
    editingResource,
    setEditingResource,
    editingDataResource,
    setEditingDataResource,
    confirmAction,
    setConfirmAction,
    confirmRollout,
    setConfirmRollout,
    createdJob,
    setCreatedJob,
    handleDelete,
    handleRolloutDeployment,
    refetch,
    menuOpen,
    setMenuOpen
}) => {
    return (
        <>
            {/* YAML Editor Modal */}
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        refetch();
                    }}
                />
            )}

            {/* Menu overlay */}
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}

            {/* Delete confirmation modal */}
            <DeleteConfirmationModal
                confirmAction={confirmAction}
                setConfirmAction={setConfirmAction}
                handleDelete={handleDelete}
            />

            {/* Rollout confirmation modal */}
            <RolloutConfirmationModal
                confirmRollout={confirmRollout}
                setConfirmRollout={setConfirmRollout}
                handleRolloutDeployment={handleRolloutDeployment}
            />

            {/* Job created success modal */}
            <JobCreatedModal
                createdJob={createdJob}
                setCreatedJob={setCreatedJob}
            />

            {/* Data Editor Modal - Edit In Place */}
            {editingDataResource && (
                <DataEditor
                    resource={editingDataResource}
                    data={editingDataResource.details?.data || {}}
                    isSecret={editingDataResource.kind === 'Secret'}
                    onClose={() => setEditingDataResource(null)}
                    onSaved={() => {
                        setEditingDataResource(null);
                        refetch();
                    }}
                />
            )}
        </>
    );
};

export default WorkloadModals;
