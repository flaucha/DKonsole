import React from 'react';
import { Link } from 'react-router-dom';

const JobCreatedModal = ({ createdJob, setCreatedJob }) => {
    if (!createdJob) return null;

    return (
        <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                <h3 className="text-lg font-semibold text-white mb-2">
                    Job creado
                </h3>
                <p className="text-sm text-gray-300 mb-4">
                    Se ha creado el job{' '}
                    <Link
                        to={`/dashboard/workloads/Job?search=${createdJob.name}&namespace=${createdJob.namespace}`}
                        onClick={() => setCreatedJob(null)}
                        className="text-blue-400 hover:text-blue-300 hover:underline font-medium"
                    >
                        {createdJob.name}
                    </Link>
                </p>
                <div className="flex justify-end space-x-3">
                    <button
                        onClick={() => setCreatedJob(null)}
                        className="px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-md transition-colors"
                    >
                        Aceptar
                    </button>
                </div>
            </div>
        </div>
    );
};

export default JobCreatedModal;
