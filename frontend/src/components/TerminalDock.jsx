import React, { useMemo, useRef, useState, useEffect } from 'react';
import { Terminal, PinOff, X, ChevronLeft, ChevronRight, Eye, EyeOff } from 'lucide-react';
import TerminalViewerInline from './details/TerminalViewerInline';
import { useTerminalDock } from '../context/TerminalDockContext';

const SessionChip = ({ session, isActive, onSelect, onUnpin, onClose }) => (
    <div
        onClick={onSelect}
        className={`group flex items-center gap-3 px-3 py-2 rounded-lg border cursor-pointer min-w-[200px] transition-all duration-200 ${
            isActive
                ? 'bg-gray-800 border-blue-500 shadow-lg shadow-blue-900/30'
                : 'bg-gray-800/50 border-gray-700 hover:border-gray-500 hover:bg-gray-800/80'
        }`}
    >
        <div className="flex items-center gap-2 truncate">
            <div className="flex flex-col leading-tight truncate">
                <div className="flex items-center gap-2">
                    <Terminal size={14} className="text-gray-300 shrink-0" />
                    <span className="text-xs font-semibold text-gray-100 truncate">{session.podName}</span>
                </div>
                <div className="text-[11px] text-gray-400 truncate">{session.container || '—'}</div>
            </div>
        </div>
        <div className="flex items-center gap-1 ml-auto shrink-0">
            <button
                onClick={(e) => {
                    e.stopPropagation();
                    onUnpin();
                }}
                className="p-1 rounded-md border text-gray-300 transition-colors bg-gray-800 border-gray-700 hover:border-gray-500 hover:text-white"
                title="Unpin session"
            >
                <PinOff size={14} />
            </button>
            <button
                onClick={(e) => {
                    e.stopPropagation();
                    onClose();
                }}
                className="p-1 rounded-md border border-gray-700 text-gray-400 hover:text-red-300 hover:border-red-700 hover:bg-red-900/40 transition-colors"
                title="Close session"
            >
                <X size={14} />
            </button>
        </div>
    </div>
);

const TerminalDock = () => {
    const { sessions, activeId, setActiveId, removeSession } = useTerminalDock();
    const [canScrollLeft, setCanScrollLeft] = useState(false);
    const [canScrollRight, setCanScrollRight] = useState(false);
    const [collapsed, setCollapsed] = useState(() => {
        if (typeof localStorage === 'undefined') return false;
        const saved = localStorage.getItem('dkonsole_dock_collapsed');
        return saved === 'true';
    });
    const scrollRef = useRef(null);

    const activeSession = useMemo(
        () => sessions.find(s => s.id === activeId) || sessions[0],
        [sessions, activeId]
    );

    const updateScrollState = () => {
        const el = scrollRef.current;
        if (!el) return;
        const maxScrollLeft = el.scrollWidth - el.clientWidth;
        setCanScrollLeft(el.scrollLeft > 4);
        setCanScrollRight(el.scrollLeft < maxScrollLeft - 4);
    };

    useEffect(() => {
        updateScrollState();
        const el = scrollRef.current;
        const handleResize = () => updateScrollState();
        if (el) {
            el.addEventListener('scroll', updateScrollState);
        }
        window.addEventListener('resize', handleResize);

        return () => {
            if (el) {
                el.removeEventListener('scroll', updateScrollState);
            }
            window.removeEventListener('resize', handleResize);
        };
    }, [sessions.length]);

    useEffect(() => {
        try {
            localStorage.setItem('dkonsole_dock_collapsed', collapsed ? 'true' : 'false');
        } catch {
            // ignore storage errors
        }
    }, [collapsed]);

    if (!sessions.length) return null;

    const scroll = (direction) => {
        const el = scrollRef.current;
        if (!el) return;
        const delta = Math.max(el.clientWidth * 0.6, 240);
        el.scrollBy({ left: direction === 'left' ? -delta : delta, behavior: 'smooth' });
    };

    return (
        <>
            {/* Dock bar inside header */}
            <div className="bg-gray-900 border-t border-b border-gray-800 px-4 py-2">
                <div className="flex items-center gap-3">
                    <div className="flex items-center gap-2 text-xs uppercase tracking-wide text-gray-400">
                        <Terminal size={14} className="text-gray-300" />
                        <span>Terminals fijados</span>
                    </div>
                    <button
                        onClick={() => setCollapsed(!collapsed)}
                        className="px-2 py-1 rounded-md border text-xs transition-colors bg-gray-800 border-gray-700 text-gray-300 hover:text-white hover:border-gray-500"
                        title={collapsed ? 'Mostrar terminal' : 'Ocultar terminal'}
                    >
                        <span className="flex items-center gap-1">
                            {collapsed ? <Eye size={14} /> : <EyeOff size={14} />}
                            {collapsed ? 'Mostrar' : 'Ocultar'}
                        </span>
                    </button>
                    <div className="relative flex-1 overflow-hidden">
                        {canScrollLeft && (
                            <button
                                onClick={() => scroll('left')}
                                className="absolute left-0 top-1/2 -translate-y-1/2 z-10 p-1 rounded-full bg-gray-800 border border-gray-700 text-gray-300 hover:text-white hover:border-gray-500 shadow"
                                title="Scroll left"
                            >
                                <ChevronLeft size={16} />
                            </button>
                        )}
                        <div
                            ref={scrollRef}
                            className="flex gap-2 overflow-x-auto pb-1 px-6"
                        >
                            {sessions.map(session => (
                                <SessionChip
                                    key={session.id}
                                    session={session}
                                    isActive={activeSession?.id === session.id}
                                    onSelect={() => setActiveId(session.id)}
                                    onUnpin={() => removeSession(session.id)}
                                    onClose={() => removeSession(session.id)}
                                />
                            ))}
                        </div>
                        {canScrollRight && (
                            <button
                                onClick={() => scroll('right')}
                                className="absolute right-0 top-1/2 -translate-y-1/2 z-10 p-1 rounded-full bg-gray-800 border border-gray-700 text-gray-300 hover:text-white hover:border-gray-500 shadow"
                                title="Scroll right"
                            >
                                <ChevronRight size={16} />
                            </button>
                        )}
                    </div>
                </div>
            </div>

            {/* Active window floating so it survives view changes */}
            {activeSession && !collapsed && (
                <div className="fixed bottom-4 right-4 w-[760px] max-w-[95vw] h-[55vh] z-40">
                    {sessions.map(session => (
                        <TerminalViewerInline
                            key={session.id}
                            namespace={session.namespace}
                            pod={session.podName}
                            container={session.container}
                            isActive={activeSession.id === session.id}
                            onPinToggle={() => removeSession(session.id)}
                            onClose={() => removeSession(session.id)}
                            pinned
                            sessionLabel={`${session.podName} / ${session.container}`}
                        />
                    ))}
                </div>
            )}
        </>
    );
};

export default TerminalDock;
