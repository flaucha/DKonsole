import React, { useState, useEffect, useRef } from 'react';
import { X, Terminal, Download, Pause, Play } from 'lucide-react';
import { useAuth } from '../context/AuthContext';
import { logger } from '../utils/logger';

const LogViewer = ({ namespace, pod, container, onClose }) => {
    const [logs, setLogs] = useState([]);
    const [isPaused, setIsPaused] = useState(false);
    const bottomRef = useRef(null);
    const streamRef = useRef(null);
    const { authFetch } = useAuth();

    useEffect(() => {
        const fetchLogs = async () => {
            try {
                const response = await authFetch(`/api/pods/logs?namespace=${namespace}&pod=${pod}&container=${container || ''}`);
                const reader = response.body.getReader();
                const decoder = new TextDecoder();
                streamRef.current = reader;

                while (true) {
                    const { value, done } = await reader.read();
                    if (done) break;
                    const text = decoder.decode(value, { stream: true });
                    setLogs(prev => [...prev, ...text.split('\n').filter(Boolean)]);
                }
            } catch (error) {
                logger.error('Error streaming logs:', error);
                setLogs(prev => [...prev, `Error: ${error.message}`]);
            }
        };

        fetchLogs();

        return () => {
            if (streamRef.current) {
                streamRef.current.cancel();
            }
        };
    }, [namespace, pod, container]);

    useEffect(() => {
        if (!isPaused && bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth' });
        }
    }, [logs, isPaused]);

    const handleDownload = () => {
        const blob = new Blob([logs.join('\n')], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${pod}-logs.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 w-full max-w-4xl h-[80vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800 rounded-t-lg">
                    <div className="flex items-center space-x-2">
                        <Terminal size={18} className="text-gray-400" />
                        <span className="font-mono text-sm text-gray-200">{pod}</span>
                        {container && <span className="text-xs text-gray-500">({container})</span>}
                    </div>
                    <div className="flex items-center space-x-2">
                        <button
                            onClick={() => setIsPaused(!isPaused)}
                            className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                            title={isPaused ? "Resume auto-scroll" : "Pause auto-scroll"}
                        >
                            {isPaused ? <Play size={16} /> : <Pause size={16} />}
                        </button>
                        <button
                            onClick={handleDownload}
                            className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                            title="Download logs"
                        >
                            <Download size={16} />
                        </button>
                        <button
                            onClick={onClose}
                            className="p-1.5 hover:bg-red-900/50 rounded text-gray-400 hover:text-red-400 transition-colors"
                            title="Close"
                        >
                            <X size={18} />
                        </button>
                    </div>
                </div>

                {/* Terminal Content */}
                <div className="flex-1 overflow-auto p-4 font-mono text-xs md:text-sm bg-black text-green-400">
                    {logs.map((line, i) => (
                        <div key={i} className="whitespace-pre-wrap break-all hover:bg-gray-900/30">
                            {line}
                        </div>
                    ))}
                    <div ref={bottomRef} />
                </div>
            </div>
        </div>
    );
};

export default LogViewer;
