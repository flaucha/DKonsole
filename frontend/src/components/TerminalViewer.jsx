import React, { useEffect, useRef } from 'react';
import { X, Terminal as TerminalIcon } from 'lucide-react';
import { Terminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

const TerminalViewer = ({ namespace, pod, container, onClose }) => {
    const termContainerRef = useRef(null);
    const termRef = useRef(null);
    const fitAddonRef = useRef(null);

    // Initialize terminal once
    useEffect(() => {
        const term = new Terminal({
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

        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        // Token is automatically sent via HttpOnly cookie, no need to pass it in URL
        const wsUrl = `${protocol}//${window.location.host}/api/pods/exec?namespace=${namespace}&pod=${pod}&container=${container || ''}`;

        const ws = new WebSocket(wsUrl);
        ws.binaryType = 'arraybuffer';

        const decoder = new TextDecoder();
        const dataListener = term.onData((data) => {
            if (ws.readyState === WebSocket.OPEN) {
                ws.send(data);
            }
        });

        ws.onopen = () => {
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

    // Refit when content changes
    useEffect(() => {
        fitAddonRef.current?.fit();
    }, [namespace, pod, container]);

    return (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
            <div className="bg-gray-900 w-full max-w-4xl h-[80vh] rounded-lg border border-gray-700 flex flex-col shadow-2xl overflow-hidden">
                {/* Header */}
                <div className="flex items-center justify-between px-4 py-3 border-b border-gray-700 bg-gray-800 rounded-t-lg">
                    <div className="flex items-center space-x-2">
                        <TerminalIcon size={18} className="text-gray-400" />
                        <span className="font-mono text-sm text-gray-200">{pod}</span>
                        {container && <span className="text-xs text-gray-500">({container})</span>}
                        <span className="text-xs text-gray-600">• Terminal</span>
                    </div>
                    <button
                        onClick={onClose}
                        className="p-1.5 hover:bg-red-900/50 rounded text-gray-400 hover:text-red-400 transition-colors"
                        title="Close"
                    >
                        <X size={18} />
                    </button>
                </div>

                {/* Terminal Surface */}
                <div className="flex-1 bg-black overflow-hidden" style={{ backgroundColor: '#000000' }}>
                    <div ref={termContainerRef} className="w-full h-full" />
                </div>
            </div>
        </div>
    );
};

export default TerminalViewer;
