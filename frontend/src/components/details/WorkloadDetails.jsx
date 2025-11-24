import React from 'react';
import { Activity, Check, X, Layers, AlertTriangle, Clock, Pause, Network, Box } from 'lucide-react';
import { DetailRow, EditYamlButton } from './CommonDetails';

export const JobDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Active" value={details.active} icon={Activity} />
            <DetailRow label="Succeeded" value={details.succeeded} icon={Check} />
            <DetailRow label="Failed" value={details.failed} icon={X} />
            <DetailRow label="Parallelism" value={details.parallelism} icon={Layers} />
            <DetailRow label="Completions" value={details.completions} icon={Layers} />
            <DetailRow label="Backoff Limit" value={details.backoffLimit} icon={AlertTriangle} />
        </div>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const CronJobDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Schedule" value={details.schedule} icon={Clock} />
            <DetailRow label="Suspend" value={String(details.suspend)} icon={Pause} />
            <DetailRow label="Concurrency" value={details.concurrency} icon={Layers} />
            <DetailRow label="Start Deadline" value={details.startingDeadline} icon={Clock} />
            <DetailRow label="Last Schedule" value={details.lastSchedule} icon={Clock} />
        </div>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const StatefulSetDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Replicas" value={`${details.ready}/${details.replicas}`} icon={Layers} />
            <DetailRow label="Current" value={details.current} icon={Activity} />
            <DetailRow label="Updated" value={details.update} icon={Activity} />
            <DetailRow label="Service" value={details.serviceName} icon={Network} />
            <DetailRow label="Pod Mgmt" value={details.podManagement} icon={Box} />
            <DetailRow label="Update Strategy" value={details.updateStrategy?.type} icon={Layers} />
        </div>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const DaemonSetDetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Desired" value={details.desired} icon={Activity} />
            <DetailRow label="Current" value={details.current} icon={Activity} />
            <DetailRow label="Ready" value={details.ready} icon={Activity} />
            <DetailRow label="Available" value={details.available} icon={Check} />
            <DetailRow label="Updated" value={details.updated} icon={Layers} />
            <DetailRow label="Misscheduled" value={details.misscheduled} icon={AlertTriangle} />
        </div>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);

export const HPADetails = ({ details, onEditYAML }) => (
    <div className="p-4 bg-gray-900/50 rounded-md mt-2">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <DetailRow label="Min Replicas" value={details.minReplicas} icon={Layers} />
            <DetailRow label="Max Replicas" value={details.maxReplicas} icon={Layers} />
            <DetailRow label="Current" value={details.current} icon={Activity} />
            <DetailRow label="Desired" value={details.desired} icon={Activity} />
            <DetailRow label="Metrics" value={details.metrics ? details.metrics.map((m) => m.type).join(', ') : ''} icon={Activity} />
            <DetailRow label="Last Scale" value={details.lastScaleTime} icon={Clock} />
        </div>
        <div className="flex justify-end mt-4">
            <EditYamlButton onClick={onEditYAML} />
        </div>
    </div>
);
