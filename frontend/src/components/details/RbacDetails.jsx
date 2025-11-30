import React from 'react';
import { Key, Database, Lock, Users } from 'lucide-react';
import { EditYamlButton } from './CommonDetails';

export const ServiceAccountDetails = ({ details, onEditYAML }) => {
    const secrets = details.secrets || [];
    const imagePullSecrets = details.imagePullSecrets || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4">
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Secrets</h4>
                {secrets.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                        {secrets.map((s, i) => (
                            <span key={i} className="px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300 flex items-center">
                                <Key size={10} className="mr-1" /> {s.name}
                            </span>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No secrets</div>
                )}
            </div>
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Image Pull Secrets</h4>
                {imagePullSecrets.length > 0 ? (
                    <div className="flex flex-wrap gap-2">
                        {imagePullSecrets.map((s, i) => (
                            <span key={i} className="px-2 py-1 bg-gray-800 border border-gray-700 rounded text-xs text-gray-300 flex items-center">
                                <Database size={10} className="mr-1" /> {s.name}
                            </span>
                        ))}
                    </div>
                ) : (
                    <div className="text-sm text-gray-500 italic">No image pull secrets</div>
                )}
            </div>
        </div>
    );
};

export const RoleDetails = ({ details, onEditYAML }) => {
    const rules = details.rules || [];

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 relative">
            <div className="absolute top-4 right-4 z-10">
                <EditYamlButton onClick={onEditYAML} />
            </div>
            <div className="pr-32">
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Policy Rules</h4>
            <div className="overflow-x-auto">
                <table className="min-w-full text-xs text-left">
                    <thead>
                        <tr className="border-b border-gray-700">
                            <th className="py-2 px-2 font-medium text-gray-400">Resources</th>
                            <th className="py-2 px-2 font-medium text-gray-400">Verbs</th>
                            <th className="py-2 px-2 font-medium text-gray-400">API Groups</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {rules.map((rule, i) => (
                            <tr key={i}>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.resources || []).join(', ') || '*'}
                                </td>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.verbs || []).join(', ') || '*'}
                                </td>
                                <td className="py-2 px-2 text-gray-300">
                                    {(rule.apiGroups || []).join(', ') || '""'}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>
            </div>
        </div>
    );
};

export const BindingDetails = ({ details, onEditYAML }) => {
    const subjects = details.subjects || [];
    const roleRef = details.roleRef || {};

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2 space-y-4 relative">
            <div className="absolute top-4 right-4 z-10">
                <EditYamlButton onClick={onEditYAML} />
            </div>
            <div className="pr-32">
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Role Reference</h4>
                <div className="flex items-center space-x-2 text-sm text-gray-300">
                    <Lock size={14} className="text-gray-500" />
                    <span className="font-medium">{roleRef.kind}:</span>
                    <span>{roleRef.name}</span>
                </div>
            </div>
            <div>
                <h4 className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Subjects</h4>
                <div className="overflow-x-auto">
                    <table className="min-w-full text-xs text-left">
                        <thead>
                            <tr className="border-b border-gray-700">
                                <th className="py-2 px-2 font-medium text-gray-400">Kind</th>
                                <th className="py-2 px-2 font-medium text-gray-400">Name</th>
                                <th className="py-2 px-2 font-medium text-gray-400">Namespace</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-gray-800">
                            {subjects.map((sub, i) => (
                                <tr key={i}>
                                    <td className="py-2 px-2 text-gray-300">{sub.kind}</td>
                                    <td className="py-2 px-2 text-gray-300">{sub.name}</td>
                                    <td className="py-2 px-2 text-gray-300">{sub.namespace || '—'}</td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
            </div>
        </div>
    );
};
