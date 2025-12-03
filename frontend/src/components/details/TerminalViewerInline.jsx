import React, { useEffect, useRef } from 'react';
import { Terminal, Pin, PinOff, X, Minus } from 'lucide-react';
import { Terminal as XTerminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

const TerminalViewerInline = ({
    namespace,
    pod,
    container,
    isActive = true,
    onPinToggle,
    pinned = false,
    onClose,
    sessionLabel,
    onMinimize,
}) => {
    const termContainerRef = useRef(null);
    const termRef = useRef(null);
    const fitAddonRef = useRef(null);
    const wsRef = useRef(null);
    const containerRef = useRef(null);
    const isActiveRef = useRef(isActive);

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

    // Track active state without re-subscribing the websocket
    useEffect(() => {
        isActiveRef.current = isActive;
        if (isActive) {
            requestAnimationFrame(() => {
                if (fitAddonRef.current) {
                    fitAddonRef.current.fit();
                }
                termRef.current?.focus();
            });
        }
    }, [isActive]);

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
            if (isActiveRef.current) {
                term.focus();
            }
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
        if (!isActive) return;
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
    }, [namespace, pod, container, isActive]);

    return (
        <div
            ref={containerRef}
            className={`${isActive ? 'flex' : 'hidden'} bg-gray-900 border border-gray-700 rounded-lg flex-col h-full overflow-hidden shadow-lg`}
        >
            {/* Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-700 bg-gradient-to-r from-gray-900 to-gray-800">
                <div className="flex items-center space-x-3">
                    <div className="flex items-center space-x-1">
                        <span className="w-2.5 h-2.5 rounded-full bg-red-500/70 shadow-sm" />
                        <span className="w-2.5 h-2.5 rounded-full bg-yellow-400/70 shadow-sm" />
                        <span className="w-2.5 h-2.5 rounded-full bg-green-500/70 shadow-sm" />
                    </div>
                    <Terminal size={16} className="text-gray-300" />
                    <div className="flex flex-col leading-tight">
                        <span className="font-mono text-xs text-gray-100">{sessionLabel || pod}</span>
                        {container && <span className="text-[11px] text-gray-400">{container}</span>}
                    </div>
                    <span className="text-[11px] px-2 py-0.5 rounded-full bg-gray-800 text-gray-300 border border-gray-700">WebSocket</span>
                </div>
                <div className="flex items-center space-x-1">
                    {onPinToggle && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onPinToggle();
                            }}
                            className={`p-1.5 rounded-md border transition-colors ${pinned
                                ? 'bg-amber-900/50 border-amber-700 text-amber-300 hover:bg-amber-800/60'
                                : 'bg-gray-800 border-gray-700 text-gray-300 hover:border-gray-500 hover:text-white'
                            }`}
                            title={pinned ? 'Unpin terminal' : 'Pin terminal'}
                        >
                            {pinned ? <PinOff size={15} /> : <Pin size={15} />}
                        </button>
                    )}
                    {onMinimize && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onMinimize();
                            }}
                            className="p-1.5 rounded-md border border-gray-700 text-gray-400 hover:text-yellow-200 hover:border-yellow-700 hover:bg-yellow-900/30 transition-colors"
                            title="Minimize terminal"
                        >
                            <Minus size={15} />
                        </button>
                    )}
                    {onClose && (
                        <button
                            onClick={(e) => {
                                e.stopPropagation();
                                onClose();
                            }}
                            className="p-1.5 rounded-md border border-gray-700 text-gray-400 hover:text-red-300 hover:border-red-700 hover:bg-red-900/40 transition-colors"
                            title="Close terminal"
                        >
                            <X size={15} />
                        </button>
                    )}
                </div>
            </div>

            {/* Terminal Surface */}
            <div className="flex-1 bg-black overflow-hidden">
                <div ref={termContainerRef} className="w-full h-full" />
            </div>
        </div>
    );
};

export default TerminalViewerInline;
