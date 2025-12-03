import React, { useMemo, useRef, useState, useEffect } from 'react';
import { Terminal, PinOff, X, ChevronLeft, ChevronRight } from 'lucide-react';
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
    const scrollRef = useRef(null);

    const activeSession = useMemo(
        () => sessions.find(s => s.id === activeId),
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

    if (!sessions.length) return null;

    const scroll = (direction) => {
        const el = scrollRef.current;
        if (!el) return;
        const delta = Math.max(el.clientWidth * 0.6, 240);
        el.scrollBy({ left: direction === 'left' ? -delta : delta, behavior: 'smooth' });
    };

    return (
        <>
            <div className="flex items-center gap-2 w-full min-w-0">
                <div className="flex items-center gap-2 text-xs uppercase tracking-wide text-gray-400 shrink-0">
                    <Terminal size={14} className="text-gray-300" />
                </div>
                <div className="relative flex-1 min-w-0">
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
                                onSelect={() => {
                                    setActiveId(session.id);
                                }}
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

            <div className={`fixed bottom-2 right-2 w-[900px] max-w-[98vw] h-[70vh] z-40 pointer-events-none flex items-end justify-end ${activeSession ? '' : 'hidden'}`}>
                {sessions.map(session => (
                    <div
                        key={session.id}
                        className="pointer-events-auto w-[900px] max-w-[98vw] h-[70vh] flex items-end justify-end"
                    >
                        <TerminalViewerInline
                            namespace={session.namespace}
                            pod={session.podName}
                            container={session.container}
                            isActive={activeId === session.id}
                            onPinToggle={() => removeSession(session.id)}
                            onClose={() => removeSession(session.id)}
                            onMinimize={() => setActiveId(null)}
                            pinned
                            sessionLabel={`${session.podName} / ${session.container}`}
                        />
                    </div>
                ))}
            </div>
        </>
    );
};

export default TerminalDock;
