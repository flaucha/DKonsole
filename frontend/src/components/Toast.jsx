import React, { useState } from 'react';
import { X, CheckCircle, AlertCircle, AlertTriangle, Info } from 'lucide-react';
import { useToast } from '../context/ToastContext';

const Toast = ({ toast }) => {
    const { removeToast } = useToast();
    const [isExiting, setIsExiting] = useState(false);

    const handleClose = () => {
        setIsExiting(true);
        setTimeout(() => {
            removeToast(toast.id);
        }, 300); // Match animation duration
    };

    const typeConfig = {
        success: {
            icon: CheckCircle,
            bgClass: 'bg-green-500/10 border-green-500/50',
            textClass: 'text-green-400',
            iconClass: 'text-green-400'
        },
        error: {
            icon: AlertCircle,
            bgClass: 'bg-red-500/10 border-red-500/50',
            textClass: 'text-red-400',
            iconClass: 'text-red-400'
        },
        warning: {
            icon: AlertTriangle,
            bgClass: 'bg-yellow-500/10 border-yellow-500/50',
            textClass: 'text-yellow-400',
            iconClass: 'text-yellow-400'
        },
        info: {
            icon: Info,
            bgClass: 'bg-blue-500/10 border-blue-500/50',
            textClass: 'text-blue-400',
            iconClass: 'text-blue-400'
        }
    };

    const config = typeConfig[toast.type] || typeConfig.info;
    const Icon = config.icon;

    return (
        <div
            className={`
                flex items-start gap-3 p-4 rounded-lg border backdrop-blur-sm shadow-lg
                ${config.bgClass}
                transition-all duration-300 ease-out
                ${isExiting ? 'opacity-0 translate-x-full' : 'opacity-100 translate-x-0'}
                min-w-[320px] max-w-md
            `}
        >
            <Icon className={`flex-shrink-0 w-5 h-5 mt-0.5 ${config.iconClass}`} />
            <p className={`flex-1 text-sm ${config.textClass}`}>{toast.message}</p>
            <button
                onClick={handleClose}
                className="flex-shrink-0 p-0.5 hover:bg-white/10 rounded transition-colors"
            >
                <X className={`w-4 h-4 ${config.textClass}`} />
            </button>
        </div>
    );
};

export const ToastContainer = () => {
    const { toasts } = useToast();

    return (
        <div className="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
            {toasts.map(toast => (
                <div key={toast.id} className="pointer-events-auto">
                    <Toast toast={toast} />
                </div>
            ))}
        </div>
    );
};

export default Toast;
