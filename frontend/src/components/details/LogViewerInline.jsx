import React, { useState, useEffect, useRef } from 'react';
import { Terminal, Play, Pause, Download } from 'lucide-react';
import { useAuth } from '../../context/AuthContext';

const LogViewerInline = ({ namespace, pod, container }) => {
    const [logs, setLogs] = useState([]);
    const [isPaused, setIsPaused] = useState(false);
    const bottomRef = useRef(null);
    const streamRef = useRef(null);
    const { authFetch } = useAuth();

    useEffect(() => {
        setLogs([]); // Clear logs when container changes
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
                console.error('Error streaming logs:', error);
                setLogs(prev => [...prev, `Error: ${error.message}`]);
            }
        };

        fetchLogs();

        return () => {
            if (streamRef.current) {
                streamRef.current.cancel();
            }
        };
    }, [namespace, pod, container, authFetch]);

    const containerRef = useRef(null);

    useEffect(() => {
        if (!isPaused && bottomRef.current) {
            bottomRef.current.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
        }
    }, [logs, isPaused]);

    // Scroll into view when component mounts (like terminal does)
    useEffect(() => {
        if (containerRef.current) {
            requestAnimationFrame(() => {
                containerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            });
        }
    }, [namespace, pod, container]);

    const handleDownload = () => {
        const blob = new Blob([logs.join('\n')], { type: 'text/plain' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `${pod}-${container || 'default'}-logs.txt`;
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
    };

    const handleClear = () => {
        setLogs([]);
    };

    return (
        <div ref={containerRef} className="bg-gray-900 border border-gray-700 rounded-lg flex flex-col h-full overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-700 bg-gray-800">
                <div className="flex items-center space-x-2">
                    <Terminal size={16} className="text-gray-400" />
                    <span className="font-mono text-xs text-gray-200">{pod}</span>
                    {container && <span className="text-xs text-gray-500">({container})</span>}
                </div>
                <div className="flex items-center space-x-2">
                    <button
                        onClick={() => setIsPaused(!isPaused)}
                        className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                        title={isPaused ? "Resume auto-scroll" : "Pause auto-scroll"}
                    >
                        {isPaused ? <Play size={14} /> : <Pause size={14} />}
                    </button>
                    <button
                        onClick={handleClear}
                        className="px-2 py-1 text-xs bg-gray-800 hover:bg-gray-700 text-gray-300 rounded border border-gray-700 transition-colors"
                        title="Clear logs"
                    >
                        Clear
                    </button>
                    <button
                        onClick={handleDownload}
                        className="p-1.5 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                        title="Download logs"
                    >
                        <Download size={14} />
                    </button>
                </div>
            </div>

            {/* Logs Content */}
            <div className="flex-1 overflow-auto p-4 font-mono text-xs bg-black text-green-400">
                {logs.length === 0 ? (
                    <div className="text-gray-500 italic">Loading logs...</div>
                ) : (
                    <>
                        {logs.map((line, i) => (
                            <div key={i} className="whitespace-pre-wrap break-all hover:bg-gray-900/30 py-0.5">
                                {line}
                            </div>
                        ))}
                        <div ref={bottomRef} />
                    </>
                )}
            </div>
        </div>
    );
};

export default LogViewerInline;
