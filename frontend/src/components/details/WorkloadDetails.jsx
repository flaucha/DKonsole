import React, { useState } from 'react';
import { Activity, Check, X, Layers, AlertTriangle, Clock, Pause, Network, Box } from 'lucide-react';
import { DetailRow } from './CommonDetails';
import AssociatedPods from './AssociatedPods';

const TabButton = ({ active, label, onClick }) => (
    <button
        onClick={onClick}
        className={`px-4 py-1.5 rounded text-sm font-medium transition-colors ${active
            ? 'bg-gray-700 text-white shadow-sm'
            : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800'
            }`}
    >
        {label}
    </button>
);

export const JobDetails = ({ details, namespace }) => {
    const [activeTab, setActiveTab] = useState('details');
    // Job selector can be complex, usually MatchLabels is what we want
    const selector = details.podSelector?.matchLabels || details.podSelector;

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <TabButton active={activeTab === 'details'} label="Details" onClick={() => setActiveTab('details')} />
                <TabButton active={activeTab === 'runs'} label="Runs" onClick={() => setActiveTab('runs')} />
            </div>

            {activeTab === 'details' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <DetailRow label="Active" value={details.active} icon={Activity} />
                        <DetailRow label="Succeeded" value={details.succeeded} icon={Check} />
                        <DetailRow label="Failed" value={details.failed} icon={X} />
                        <DetailRow label="Parallelism" value={details.parallelism} icon={Layers} />
                        <DetailRow label="Completions" value={details.completions} icon={Layers} />
                        <DetailRow label="Backoff Limit" value={details.backoffLimit} icon={AlertTriangle} />
                    </div>

                </>
            )}

            {activeTab === 'runs' && (
                <AssociatedPods namespace={namespace} selector={selector} />
            )}
        </div>
    );
};

export const CronJobDetails = ({ details }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Schedule" value={details.schedule} icon={Clock} />
            <DetailRow label="Suspend" value={String(details.suspend)} icon={Pause} />
            <DetailRow label="Concurrency" value={details.concurrency} icon={Layers} />
            <DetailRow label="Start Deadline" value={details.startingDeadline} icon={Clock} />
            <DetailRow label="Last Schedule" value={details.lastSchedule} icon={Clock} />
        </div>

    </div>
);

export const StatefulSetDetails = ({ details, namespace }) => {
    const [activeTab, setActiveTab] = useState('details');

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <TabButton active={activeTab === 'details'} label="Details" onClick={() => setActiveTab('details')} />
                <TabButton active={activeTab === 'pods'} label="Pod List" onClick={() => setActiveTab('pods')} />
            </div>

            {activeTab === 'details' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <DetailRow label="Replicas" value={`${details.ready}/${details.replicas}`} icon={Layers} />
                        <DetailRow label="Current" value={details.current} icon={Activity} />
                        <DetailRow label="Updated" value={details.update} icon={Activity} />
                        <DetailRow label="Service" value={details.serviceName} icon={Network} />
                        <DetailRow label="Pod Mgmt" value={details.podManagement} icon={Box} />
                        <DetailRow label="Update Strategy" value={details.updateStrategy?.type} icon={Layers} />
                    </div>

                </>
            )}

            {activeTab === 'pods' && (
                <AssociatedPods namespace={namespace} selector={details.selector} />
            )}
        </div>
    );
};

export const DaemonSetDetails = ({ details, namespace }) => {
    const [activeTab, setActiveTab] = useState('details');

    return (
        <div className="p-4 bg-gray-900/50 rounded-md mt-2">
            <div className="flex space-x-1 bg-gray-800/50 p-1 rounded-md mb-4 w-fit">
                <TabButton active={activeTab === 'details'} label="Details" onClick={() => setActiveTab('details')} />
                <TabButton active={activeTab === 'pods'} label="Pod List" onClick={() => setActiveTab('pods')} />
            </div>

            {activeTab === 'details' && (
                <>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                        <DetailRow label="Desired" value={details.desired} icon={Activity} />
                        <DetailRow label="Current" value={details.current} icon={Activity} />
                        <DetailRow label="Ready" value={details.ready} icon={Activity} />
                        <DetailRow label="Available" value={details.available} icon={Check} />
                        <DetailRow label="Updated" value={details.updated} icon={Layers} />
                        <DetailRow label="Misscheduled" value={details.misscheduled} icon={AlertTriangle} />
                    </div>

                </>
            )}

            {activeTab === 'pods' && (
                <AssociatedPods namespace={namespace} selector={details.selector} />
            )}
        </div>
    );
};

export const HPADetails = ({ details }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Min Replicas" value={details.minReplicas} icon={Layers} />
            <DetailRow label="Max Replicas" value={details.maxReplicas} icon={Layers} />
            <DetailRow label="Current" value={details.current} icon={Activity} />
            <DetailRow label="Desired" value={details.desired} icon={Activity} />
            <DetailRow label="Metrics" value={details.metrics ? details.metrics.map((m) => m.type).join(', ') : ''} icon={Activity} />
            <DetailRow label="Last Scale" value={details.lastScaleTime} icon={Clock} />
        </div>

    </div>
);
