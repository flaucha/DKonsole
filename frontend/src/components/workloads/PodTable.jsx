import React, { useState, useEffect, useRef } from 'react';
import {
    Box,
    FileText,
    Activity,
    Network,
    Server,
    HardDrive,
    Terminal,
    CirclePlus,
    CircleMinus,
    MoreVertical,
    Play,
    Pause,
    Clock,
    Download
} from 'lucide-react';
import { Terminal as XTerminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';
import YamlEditor from '../YamlEditor';
import PodMetrics from '../PodMetrics';
import { useSettings } from '../../context/SettingsContext';
import { useAuth } from '../../context/AuthContext';
import { formatDateTime, formatDateTimeShort } from '../../utils/dateUtils';
import { getExpandableRowClasses, getExpandableCellClasses, getExpandableRowStyles, getExpandableRowRowClasses } from '../../utils/expandableRow';
import { getStatusBadgeClass } from '../../utils/statusBadge';

// Inline version of LogViewer for tabs (no popup)
const LogViewerInline = ({ namespace, pod, container }) => {
    const [logs, setLogs] = useState([]);
    const [isPaused, setIsPaused] = useState(false);
    const [textColor, setTextColor] = useState(() => {
        // Cargar color guardado desde localStorage, o usar verde por defecto
        const savedColor = localStorage.getItem('dkonsole-log-text-color');
        return savedColor || 'green';
    });
    const bottomRef = useRef(null);
    const streamRef = useRef(null);
    const { authFetch } = useAuth();

    const colorOptions = [
        { name: 'gris', value: 'gray', textClass: 'text-gray-400', bgClass: 'bg-gray-500' },
        { name: 'verde', value: 'green', textClass: 'text-green-400', bgClass: 'bg-green-500' },
        { name: 'celeste', value: 'cyan', textClass: 'text-cyan-400', bgClass: 'bg-cyan-500' },
        { name: 'amarillo', value: 'yellow', textClass: 'text-yellow-400', bgClass: 'bg-yellow-500' },
        { name: 'naranja', value: 'orange', textClass: 'text-orange-400', bgClass: 'bg-orange-500' },
        { name: 'blanco', value: 'white', textClass: 'text-white', bgClass: 'bg-white' },
    ];

    // Guardar color seleccionado en localStorage
    const handleColorChange = (colorValue) => {
        setTextColor(colorValue);
        localStorage.setItem('dkonsole-log-text-color', colorValue);
    };

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
        <div className="bg-gray-900 border border-gray-700 rounded-lg flex flex-col h-full overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-700 bg-gray-800">
                <div className="flex items-center space-x-2">
                    <Terminal size={16} className="text-gray-400" />
                    <span className="font-mono text-xs text-gray-200">{pod}</span>
                    {container && <span className="text-xs text-gray-500">({container})</span>}
                </div>
                <div className="flex items-center space-x-4">
                    {/* Color Selector */}
                    <div className="flex items-center space-x-2">
                        <span className="text-xs text-gray-400">Color:</span>
                        <div className="flex items-center gap-1.5 p-1 bg-gray-900/50 rounded-md border border-gray-700">
                            {colorOptions.map((color) => (
                                <button
                                    key={color.value}
                                    onClick={() => handleColorChange(color.value)}
                                    className={`w-6 h-6 rounded border-2 transition-all hover:scale-110 ${
                                        textColor === color.value
                                            ? 'border-white scale-110 shadow-lg shadow-white/20'
                                            : 'border-gray-600 hover:border-gray-400'
                                    } ${color.bgClass}`}
                                    title={color.name}
                                />
                            ))}
                        </div>
                    </div>
                    {/* Control Buttons */}
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
            </div>

            {/* Logs Content */}
            <div className={`flex-1 overflow-auto p-4 font-mono text-xs bg-black ${colorOptions.find(c => c.value === textColor)?.textClass || 'text-green-400'}`}>
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

// Inline version of TerminalViewer for tabs (no popup)
const TerminalViewerInline = ({ namespace, pod, container }) => {
    const termContainerRef = useRef(null);
    const termRef = useRef(null);
    const fitAddonRef = useRef(null);
    const wsRef = useRef(null);
    const containerRef = useRef(null);

    // Initialize terminal once
    useEffect(() => {
        const term = new XTerminal({
            convertEol: true,
            cursorBlink: true,
            fontSize: 13,
            fontFamily: 'Menlo, Monaco, "Cascadia Mono", "Fira Code", monospace',
            theme: {
                background: '#000000',
                foreground: '#e5e7eb',
            },
        });

        const fitAddon = new FitAddon();
        term.loadAddon(fitAddon);

        term.open(termContainerRef.current);
        requestAnimationFrame(() => fitAddon.fit());
        term.focus();

        termRef.current = term;
        fitAddonRef.current = fitAddon;

        const handleResize = () => {
            if (fitAddonRef.current) {
                fitAddonRef.current.fit();
            }
        };

        window.addEventListener('resize', handleResize);

        return () => {
            window.removeEventListener('resize', handleResize);
            term.dispose();
        };
    }, []);

    // Connect WebSocket for the active pod/container
    useEffect(() => {
        const term = termRef.current;
        if (!term) return;

        // Close existing connection
        if (wsRef.current && (wsRef.current.readyState === WebSocket.OPEN || wsRef.current.readyState === WebSocket.CONNECTING)) {
            wsRef.current.close();
        }

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/pods/exec?namespace=${namespace}&pod=${pod}&container=${container || ''}`;

        const ws = new WebSocket(wsUrl);
        ws.binaryType = 'arraybuffer';
        wsRef.current = ws;

        const decoder = new TextDecoder();
        const dataListener = term.onData((data) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            }
        });

        ws.onopen = () => {
            term.clear();
            term.writeln(`\x1b[33mConnected to ${pod}${container ? ` (${container})` : ''}\x1b[0m`);
            fitAddonRef.current?.fit();
            term.focus();
        };

        ws.onmessage = (event) => {
            if (typeof event.data === 'string') {
                term.write(event.data);
            } else if (event.data instanceof ArrayBuffer) {
                term.write(decoder.decode(event.data));
            } else if (event.data instanceof Blob) {
                event.data.arrayBuffer().then((buf) => term.write(decoder.decode(buf)));
            }
        };

        ws.onerror = () => {
            term.writeln('\r\n\x1b[31mWebSocket error. Check connection.\x1b[0m');
        };

        ws.onclose = () => {
            term.writeln('\r\n\x1b[31mConnection closed.\x1b[0m');
        };

        return () => {
            dataListener.dispose();
            if (ws.readyState === WebSocket.OPEN || ws.readyState === WebSocket.CONNECTING) {
                ws.close();
            }
        };
    }, [namespace, pod, container]);

    // Refit when content changes and scroll into view
    useEffect(() => {
        // Use requestAnimationFrame to ensure DOM is ready
        requestAnimationFrame(() => {
            if (fitAddonRef.current) {
                fitAddonRef.current.fit();
            }
            // Scroll the terminal container into view
            if (containerRef.current) {
                containerRef.current.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }
        });
    }, [namespace, pod, container]);

    return (
        <div ref={containerRef} className="bg-gray-900 border border-gray-700 rounded-lg flex flex-col h-full overflow-hidden">
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-700 bg-gray-800">
                <div className="flex items-center space-x-2">
                    <Terminal size={16} className="text-gray-400" />
                    <span className="font-mono text-xs text-gray-200">{pod}</span>
                    {container && <span className="text-xs text-gray-500">({container})</span>}
                    <span className="text-xs text-gray-600">• Terminal</span>
                </div>
            </div>

            {/* Terminal Surface */}
            <div className="flex-1 bg-black overflow-hidden">
                <div ref={termContainerRef} className="w-full h-full" />
            </div>
        </div>
    );
};

const DetailRow = ({ label, value, icon: Icon, children }) => (
    <div className="flex items-center justify-between bg-gray-800 px-3 py-2 rounded border border-gray-700 mb-2">
        <div className="flex items-center">
            {Icon && <Icon size={14} className="mr-2 text-gray-500" />}
            <span className="text-xs text-gray-400">{label}</span>
        </div>
        <div className="flex items-center">
            <span className="text-sm font-mono text-white break-all text-right">
                {Array.isArray(value) ? (
                    value.length > 0 ? value.join(', ') : <span className="text-gray-600 italic">None</span>
                ) : (
                    value || <span className="text-gray-600 italic">None</span>
                )}
            </span>
            {children}
        </div>
    </div>
);

const EditYamlButton = ({ onClick }) => (
    <button
        onClick={onClick}
        className="flex items-center px-2.5 py-1 bg-gray-800 hover:bg-gray-700 text-gray-300 rounded-md border border-gray-600 text-xs transition-colors"
    >
        <FileText size={12} className="mr-1.5" />
        Edit YAML
    </button>
);

const PodDetails = ({ details, onEditYAML, pod }) => {
    const { authFetch } = useAuth();
    const [activeTab, setActiveTab] = useState('details');
    const [selectedContainer, setSelectedContainer] = useState(null);
    const [events, setEvents] = useState([]);
    const [loadingEvents, setLoadingEvents] = useState(false);
    const containers = details.containers || [];
    const metrics = details.metrics || {};
    const terminalContainerRef = useRef(null);

    // Set default container when component mounts or containers change
    useEffect(() => {
        if (containers.length > 0 && !selectedContainer) {
            setSelectedContainer(containers[0]);
        }
    }, [containers, selectedContainer]);

    // Scroll into view when terminal tab is activated
    useEffect(() => {
        if (activeTab === 'terminal' && terminalContainerRef.current) {
            // Use setTimeout to ensure DOM is updated
            setTimeout(() => {
                terminalContainerRef.current?.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
            }, 100);
        }
    }, [activeTab]);

    // Fetch events when events tab is activated
    useEffect(() => {
        if (activeTab === 'events' && pod && !loadingEvents && events.length === 0) {
            setLoadingEvents(true);
            authFetch(`/api/pods/events?namespace=${pod.namespace}&pod=${pod.name}`)
                .then(res => res.json())
                .then(data => {
                    setEvents(data || []);
                    setLoadingEvents(false);
                })
                .catch(err => {
                    console.error('Error fetching events:', err);
                    setEvents([]);
                    setLoadingEvents(false);
                });
        }
    }, [activeTab, pod, authFetch, loadingEvents, events.length]);

    // Determine if we should use full height (for logs/terminal tabs)
    const isFullHeightTab = activeTab === 'logs' || activeTab === 'terminal';
    // Calculate height: viewport height minus header, tabs, and padding (approximately)
    const fullHeight = 'calc(100vh - 300px)';

    return (
        <div className="mt-2 flex flex-col" style={{ height: isFullHeightTab ? fullHeight : 'auto' }}>
            <div className="flex space-x-4 border-b border-gray-700 mb-4 flex-shrink-0">
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'details' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('details')}
                >
                    Details
                </button>
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'logs' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('logs')}
                >
                    Logs
                </button>
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'terminal' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('terminal')}
                >
                    Terminal
                </button>
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'metrics' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('metrics')}
                >
                    Metrics
                </button>
                <button
                    className={`pb-2 text-sm font-medium transition-colors ${activeTab === 'events' ? 'text-blue-400 border-b-2 border-blue-400' : 'text-gray-400 hover:text-gray-300'}`}
                    onClick={() => setActiveTab('events')}
                >
                    Events
                </button>
            </div>

            <div className={`transition-all duration-300 ease-in-out flex-1 flex flex-col ${isFullHeightTab ? 'min-h-0' : ''}`}>
                {activeTab === 'details' ? (
                    <div className="p-4 bg-gray-900/50 rounded-md animate-fadeIn">
                        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-3">
                            <div>
                                <DetailRow label="Node" value={details.node} icon={Server} />
                                <DetailRow label="IP" value={details.ip} icon={Network} />
                            </div>
                            <div>
                                <DetailRow label="Restarts" value={details.restarts} icon={Activity} />
                                <DetailRow label="Containers" value={containers} icon={Box} />
                            </div>
                            <div>
                                {metrics.cpu && <DetailRow label="CPU" value={metrics.cpu} icon={Activity} />}
                                {metrics.memory && <DetailRow label="Memory" value={metrics.memory} icon={HardDrive} />}
                            </div>
                        </div>
                        <div className="flex justify-end space-x-2">
                            <EditYamlButton onClick={onEditYAML} />
                        </div>
                    </div>
                ) : activeTab === 'logs' ? (
                    <div className="animate-fadeIn flex-1 flex flex-col min-h-0">
                        {containers.length > 1 && (
                            <div className="mb-3 flex items-center space-x-2 flex-shrink-0">
                                <label className="text-xs text-gray-400">Container:</label>
                                <select
                                    value={selectedContainer || ''}
                                    onChange={(e) => setSelectedContainer(e.target.value)}
                                    className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                                >
                                    {containers.map(c => (
                                        <option key={c} value={c}>{c}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {selectedContainer && pod && (
                            <div className="flex-1 min-h-0">
                                <LogViewerInline
                                    namespace={pod.namespace}
                                    pod={pod.name}
                                    container={selectedContainer}
                                />
                            </div>
                        )}
                    </div>
                ) : activeTab === 'terminal' ? (
                    <div ref={terminalContainerRef} className="animate-fadeIn flex-1 flex flex-col min-h-0">
                        {containers.length > 1 && (
                            <div className="mb-3 flex items-center space-x-2 flex-shrink-0">
                                <label className="text-xs text-gray-400">Container:</label>
                                <select
                                    value={selectedContainer || ''}
                                    onChange={(e) => setSelectedContainer(e.target.value)}
                                    className="px-3 py-1.5 bg-gray-800 border border-gray-700 rounded text-sm text-gray-200 focus:outline-none focus:border-blue-500"
                                >
                                    {containers.map(c => (
                                        <option key={c} value={c}>{c}</option>
                                    ))}
                                </select>
                            </div>
                        )}
                        {selectedContainer && pod && (
                            <div className="flex-1 min-h-0">
                                <TerminalViewerInline
                                    namespace={pod.namespace}
                                    pod={pod.name}
                                    container={selectedContainer}
                                />
                            </div>
                        )}
                    </div>
                ) : activeTab === 'metrics' ? (
                    <div className="animate-fadeIn">
                        {pod && <PodMetrics pod={{ name: pod.name }} namespace={pod.namespace} />}
                    </div>
                ) : (
                    <div className="animate-fadeIn p-4 bg-gray-900/50 rounded-md">
                        <div className="space-y-4">
                            {/* Events Section */}
                            <div>
                                <h4 className="text-sm font-semibold text-gray-300 mb-3 flex items-center">
                                    <Activity size={16} className="mr-2" />
                                    Pod Events
                                </h4>
                                {loadingEvents ? (
                                    <div className="text-gray-500 italic">Loading events...</div>
                                ) : events.length === 0 ? (
                                    <div className="text-gray-500 italic">No events available</div>
                                ) : (
                                    <div className="space-y-2 max-h-96 overflow-y-auto">
                                        {events.map((event, idx) => (
                                            <div key={idx} className={`p-3 rounded-md border ${event.type === 'Warning' ? 'bg-yellow-900/20 border-yellow-700' : 'bg-blue-900/20 border-blue-700'}`}>
                                                <div className="flex items-start justify-between">
                                                    <div className="flex-1">
                                                        <div className="flex items-center space-x-2 mb-1">
                                                            <span className={`text-xs font-semibold ${event.type === 'Warning' ? 'text-yellow-400' : 'text-blue-400'}`}>
                                                                {event.type}
                                                            </span>
                                                            <span className="text-xs text-gray-400">•</span>
                                                            <span className="text-xs font-medium text-gray-300">{event.reason}</span>
                                                            {event.count > 1 && (
                                                                <>
                                                                    <span className="text-xs text-gray-400">•</span>
                                                                    <span className="text-xs text-gray-500">x{event.count}</span>
                                                                </>
                                                            )}
                                                        </div>
                                                        <p className="text-xs text-gray-300 mb-1">{event.message}</p>
                                                        {event.source && (
                                                            <p className="text-xs text-gray-500">Source: {event.source}</p>
                                                        )}
                                                    </div>
                                                    <div className="text-xs text-gray-500 ml-4 text-right">
                                                        <div>{formatDateTime(event.lastSeen)}</div>
                                                        {event.firstSeen !== event.lastSeen && (
                                                            <div className="text-gray-600 mt-1">
                                                                First seen: {formatDateTimeShort(event.firstSeen)}
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            </div>
                                        ))}
                                    </div>
                                )}
                            </div>

                            {/* Container Status Timeline */}
                            <div>
                                <h4 className="text-sm font-semibold text-gray-300 mb-3 flex items-center">
                                    <Clock size={16} className="mr-2" />
                                    Container Status Timeline
                                </h4>
                                {details.containerStatuses && details.containerStatuses.length > 0 ? (
                                    <div className="space-y-3">
                                        {details.containerStatuses.map((containerStatus, idx) => (
                                            <div key={idx} className="p-3 bg-gray-800 rounded-md border border-gray-700">
                                                <div className="flex items-center justify-between mb-2">
                                                    <div className="flex items-center space-x-2">
                                                        <span className="text-sm font-medium text-white">{containerStatus.name}</span>
                                                        <span className={`px-2 py-0.5 text-xs rounded ${containerStatus.ready ? 'bg-green-900/30 text-green-400' : 'bg-red-900/30 text-red-400'}`}>
                                                            {containerStatus.ready ? 'Ready' : 'Not Ready'}
                                                        </span>
                                                        <span className={`px-2 py-0.5 text-xs rounded ${
                                                            containerStatus.state === 'Running' ? 'bg-blue-900/30 text-blue-400' :
                                                            containerStatus.state === 'Waiting' ? 'bg-yellow-900/30 text-yellow-400' :
                                                            containerStatus.state === 'Terminated' ? 'bg-gray-700 text-gray-400' :
                                                            'bg-gray-700 text-gray-400'
                                                        }`}>
                                                            {containerStatus.state || 'Unknown'}
                                                        </span>
                                                    </div>
                                                    <div className="text-xs text-gray-500">
                                                        Restarts: {containerStatus.restartCount || 0}
                                                    </div>
                                                </div>
                                                
                                                {containerStatus.state === 'Waiting' && containerStatus.reason && (
                                                    <div className="mt-2 text-xs text-yellow-300">
                                                        <span className="font-medium">Reason:</span> {containerStatus.reason}
                                                        {containerStatus.message && (
                                                            <span className="block mt-1 text-gray-400">{containerStatus.message}</span>
                                                        )}
                                                    </div>
                                                )}
                                                
                                                {containerStatus.state === 'Running' && containerStatus.startedAt && (
                                                    <div className="mt-2 text-xs text-gray-400">
                                                        Started: {formatDateTime(containerStatus.startedAt)}
                                                    </div>
                                                )}
                                                
                                                {containerStatus.state === 'Terminated' && (
                                                    <div className="mt-2 space-y-1">
                                                        {containerStatus.reason && (
                                                            <div className="text-xs text-red-300">
                                                                <span className="font-medium">Reason:</span> {containerStatus.reason}
                                                            </div>
                                                        )}
                                                        {containerStatus.exitCode !== undefined && (
                                                            <div className="text-xs text-gray-400">
                                                                Exit Code: {containerStatus.exitCode}
                                                            </div>
                                                        )}
                                                        {containerStatus.startedAt && (
                                                            <div className="text-xs text-gray-400">
                                                                Started: {formatDateTime(containerStatus.startedAt)}
                                                            </div>
                                                        )}
                                                        {containerStatus.finishedAt && (
                                                            <div className="text-xs text-gray-400">
                                                                Finished: {formatDateTime(containerStatus.finishedAt)}
                                                            </div>
                                                        )}
                                                    </div>
                                                )}
                                                
                                                {containerStatus.image && (
                                                    <div className="mt-2 text-xs text-gray-500">
                                                        Image: {containerStatus.image}
                                                    </div>
                                                )}
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    <div className="text-gray-500 italic">No container status information available</div>
                                )}
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </div>
    );
};

const PodTable = ({ namespace, resources, loading, onReload }) => {
    const [expandedId, setExpandedId] = useState(null);
    const [editingResource, setEditingResource] = useState(null);
    const [sortField, setSortField] = useState('name');
    const [sortDirection, setSortDirection] = useState('asc');
    const { currentCluster } = useSettings();
    const { authFetch } = useAuth();
    const [menuOpen, setMenuOpen] = useState(null);
    const [confirmAction, setConfirmAction] = useState(null);
    const [filter, setFilter] = useState('');
    const [isSearchFocused, setIsSearchFocused] = useState(false);

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

    const filteredResources = resources.filter(r => {
        if (!filter) return true;
        return r.name.toLowerCase().includes(filter.toLowerCase());
    });

    const sortedResources = [...filteredResources].sort((a, b) => {
        const dir = sortDirection === 'asc' ? 1 : -1;
        const getVal = (item) => {
            switch (sortField) {
                case 'name':
                    return item.name || '';
                case 'status':
                    return item.status || '';
                case 'created':
                    return new Date(item.created).getTime() || 0;
                case 'cpu': {
                    if (!item.details?.metrics?.cpu) return 0;
                    const cpuStr = item.details.metrics.cpu.trim();
                    if (cpuStr.endsWith('m')) return parseFloat(cpuStr.replace('m', '')) || 0;
                    const val = parseFloat(cpuStr);
                    return isNaN(val) ? 0 : val * 1000;
                }
                case 'memory': {
                    if (!item.details?.metrics?.memory) return 0;
                    const memStr = item.details.metrics.memory.toUpperCase().trim();
                    const num = parseFloat(memStr);
                    if (isNaN(num)) return 0;
                    if (memStr.includes('GI')) return num * 1024;
                    if (memStr.includes('MI')) return num;
                    if (memStr.includes('KI')) return num / 1024;
                    return num;
                }
                case 'ready': {
                    if (!item.details?.ready) return -1;
                    const readyStr = item.details.ready.toString();
                    const parts = readyStr.split('/');
                    if (parts.length === 2) {
                        const ready = parseFloat(parts[0]) || 0;
                        const total = parseFloat(parts[1]) || 0;
                        return total > 0 ? (ready / total) : -1;
                    }
                    return -1;
                }
                case 'restarts': {
                    return item.details?.restarts || 0;
                }
                default:
                    return '';
            }
        };
        const va = getVal(a);
        const vb = getVal(b);
        if (typeof va === 'number' && typeof vb === 'number') {
            return (va - vb) * dir;
        }
        return String(va).localeCompare(String(vb)) * dir;
    });

    const renderSortIndicator = (field) => {
        if (sortField !== field) return null;
        return sortDirection === 'asc' ? '↑' : '↓';
    };

    const toggleExpand = (uid) => {
        setExpandedId(current => current === uid ? null : uid);
    };

    const renderDetails = (res) => {
        const onEditYAML = () => setEditingResource(res);
        if (!res.details) return (
            <div className="p-4 text-gray-500 italic">
                No details available.
                <div className="flex justify-end mt-4">
                    <EditYamlButton onClick={onEditYAML} />
                </div>
            </div>
        );
        return (
            <PodDetails
                details={res.details}
                onEditYAML={onEditYAML}
                pod={res}
            />
        );
    };

    const triggerDelete = (res, force = false) => {
        const params = new URLSearchParams({ kind: res.kind, name: res.name });
        if (res.namespace) params.append('namespace', res.namespace);
        if (currentCluster) params.append('cluster', currentCluster);
        if (force) params.append('force', 'true');

        authFetch(`/api/resource?${params.toString()}`, { method: 'DELETE' })
            .then(async (resp) => {
                if (!resp.ok) {
                    const text = await resp.text();
                    throw new Error(text || 'Delete failed');
                }
                onReload();
            })
            .catch((err) => {
                alert(err.message || 'Failed to delete resource');
            })
            .finally(() => setConfirmAction(null));
    };

    return (
        <>
            {menuOpen && (
                <div
                    className="fixed inset-0 z-40"
                    onClick={() => setMenuOpen(null)}
                ></div>
            )}
            <div className="bg-gray-800 rounded-lg border border-gray-700 overflow-x-auto">
                <table className="min-w-full border-separate border-spacing-0">
                    <thead>
                        <tr>
                            <th className="w-8 px-2 md:px-4 py-3 bg-gray-900 rounded-tl-lg border-b border-gray-700"></th>
                            <th
                                scope="col"
                                className="px-3 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('name')}
                            >
                                Name <span className="inline-block text-[10px]">{renderSortIndicator('name')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('status')}
                            >
                                Status <span className="inline-block text-[10px]">{renderSortIndicator('status')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('ready')}
                            >
                                Ready <span className="inline-block text-[10px]">{renderSortIndicator('ready')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('restarts')}
                            >
                                Restarts <span className="inline-block text-[10px]">{renderSortIndicator('restarts')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('cpu')}
                            >
                                CPU <span className="inline-block text-[10px]">{renderSortIndicator('cpu')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-4 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('memory')}
                            >
                                Memory <span className="inline-block text-[10px]">{renderSortIndicator('memory')}</span>
                            </th>
                            <th
                                scope="col"
                                className="px-2 md:px-6 py-3 text-left text-xs font-medium text-gray-400 uppercase tracking-wider cursor-pointer select-none bg-gray-900 border-b border-gray-700"
                                onClick={() => handleSort('created')}
                            >
                                Created <span className="inline-block text-[10px]">{renderSortIndicator('created')}</span>
                            </th>
                            <th className="w-10 px-4 py-3 bg-gray-900 rounded-tr-lg border-b border-gray-700"></th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-800">
                        {sortedResources.map((res) => (
                            <React.Fragment key={res.uid}>
                                <tr
                                    onClick={() => toggleExpand(res.uid)}
                                    className={getExpandableRowRowClasses(expandedId === res.uid)}
                                >
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-gray-400 text-center">
                                        {expandedId === res.uid ? <CircleMinus size={16} /> : <CirclePlus size={16} />}
                                    </td>
                                    <td className="px-3 md:px-6 py-3 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-6 w-6 bg-gray-700 rounded flex items-center justify-center text-gray-400">
                                                <Box size={14} />
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-white">{res.name}</div>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-2 md:px-6 py-3 whitespace-nowrap">
                                        <span className={`px-2 py-1 text-xs rounded-full ${getStatusBadgeClass(res.status)}`}>
                                            {res.status}
                                        </span>
                                    </td>
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                        {res.details?.ready || '—'}
                                    </td>
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                        {res.details?.restarts || 0}
                                    </td>
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                        {res.details?.metrics?.cpu || '—'}
                                    </td>
                                    <td className="px-2 md:px-4 py-3 whitespace-nowrap text-sm text-gray-300">
                                        {res.details?.metrics?.memory || '—'}
                                    </td>
                                    <td className="px-2 md:px-6 py-3 whitespace-nowrap text-sm text-gray-400">
                                        {formatDateTimeShort(res.created)}
                                    </td>
                                    <td className="px-4 py-3 whitespace-nowrap text-gray-300" onClick={(e) => e.stopPropagation()}>
                                        <div className="relative flex items-center justify-end space-x-1">
                                            <button
                                                onClick={() => setMenuOpen(menuOpen === res.name ? null : res.name)}
                                                className="p-1 hover:bg-gray-700 rounded text-gray-400 hover:text-white transition-colors"
                                            >
                                                <MoreVertical size={16} />
                                            </button>
                                            {menuOpen === res.name && (
                                                <div className="absolute right-0 mt-1 w-36 bg-gray-800 border border-gray-700 rounded-md shadow-lg z-50">
                                                    <div className="flex flex-col">
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: false });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-gray-200 hover:bg-gray-700"
                                                        >
                                                            Delete
                                                        </button>
                                                        <button
                                                            onClick={() => {
                                                                setConfirmAction({ res, force: true });
                                                                setMenuOpen(null);
                                                            }}
                                                            className="w-full text-left px-4 py-2 text-sm text-red-300 hover:bg-red-900/40"
                                                        >
                                                            Force Delete
                                                        </button>
                                                    </div>
                                                </div>
                                            )}
                                        </div>
                                    </td>
                                </tr>
                                <tr>
                                    <td colSpan={9} className={getExpandableCellClasses(expandedId === res.uid, 9)}>
                                        <div
                                            className={getExpandableRowClasses(expandedId === res.uid, true)}
                                            style={getExpandableRowStyles(expandedId === res.uid, res.kind)}
                                        >
                                            {expandedId === res.uid && renderDetails(res)}
                                        </div>
                                    </td>
                                </tr>
                            </React.Fragment>
                        ))}
                    </tbody>
                </table>
            </div>
            {editingResource && (
                <YamlEditor
                    resource={editingResource}
                    onClose={() => setEditingResource(null)}
                    onSaved={() => {
                        setEditingResource(null);
                        onReload();
                    }}
                />
            )}
            {confirmAction && (
                <div className="fixed inset-0 bg-black/70 flex items-center justify-center z-50 p-4">
                    <div className="bg-gray-900 border border-gray-700 rounded-lg p-6 w-full max-w-md shadow-xl">
                        <h3 className="text-lg font-semibold text-white mb-2">
                            Confirm delete
                        </h3>
                        <p className="text-sm text-gray-300 mb-4">
                            {confirmAction.force ? 'Force delete' : 'Delete'} {confirmAction.res.kind} "{confirmAction.res.name}"?
                        </p>
                        <div className="flex justify-end space-x-3">
                            <button
                                onClick={() => setConfirmAction(null)}
                                className="px-4 py-2 bg-gray-800 text-gray-200 rounded-md hover:bg-gray-700 transition-colors"
                            >
                                Cancel
                            </button>
                            <button
                                onClick={() => {
                                    triggerDelete(confirmAction.res, confirmAction.force);
                                }}
                                className={`px-4 py-2 rounded-md text-white transition-colors ${confirmAction.force ? 'bg-red-700 hover:bg-red-800' : 'bg-orange-600 hover:bg-orange-700'}`}
                            >
                                {confirmAction.force ? 'Force delete' : 'Delete'}
                            </button>
                        </div>
                    </div>
                </div>
            )}
        </>
    );
};

export default PodTable;


