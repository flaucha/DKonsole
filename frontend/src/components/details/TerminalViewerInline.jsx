import React, { useEffect, useRef } from 'react';
import { Terminal } from 'lucide-react';
import { Terminal as XTerminal } from '@xterm/xterm';
import { FitAddon } from '@xterm/addon-fit';
import '@xterm/xterm/css/xterm.css';

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

export default TerminalViewerInline;
