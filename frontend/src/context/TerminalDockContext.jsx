import React, { createContext, useContext, useEffect, useMemo, useRef, useState } from 'react';

const TerminalDockContext = createContext(null);
const STORAGE_KEY = 'dkonsole_pinned_sessions';

const createSessionId = () => {
    if (typeof crypto !== 'undefined' && crypto.randomUUID) {
        return crypto.randomUUID();
    }
    return `term-${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

export const TerminalDockProvider = ({ children }) => {
    const [sessions, setSessions] = useState([]);
    const [activeId, setActiveId] = useState(null);
    const didHydrate = useRef(false);

    // Hydrate from localStorage once
    useEffect(() => {
        try {
            const raw = localStorage.getItem(STORAGE_KEY);
            if (raw) {
                const parsed = JSON.parse(raw);
                if (Array.isArray(parsed.sessions)) {
                    setSessions(parsed.sessions);
                    setActiveId(parsed.activeId || parsed.sessions[0]?.id || null);
                }
            }
        } catch {
            // ignore hydration errors
        } finally {
            didHydrate.current = true;
        }
    }, []);

    // Persist sessions + activeId
    useEffect(() => {
        if (!didHydrate.current) return;
        try {
            localStorage.setItem(STORAGE_KEY, JSON.stringify({ sessions, activeId }));
        } catch {
            // ignore persistence errors
        }
    }, [sessions, activeId]);

    // Default active when none selected
    useEffect(() => {
        if (!activeId && sessions.length) {
            setActiveId(sessions[0].id);
        }
    }, [sessions, activeId]);

    const addSession = (session) => {
        if (!session?.namespace || !session?.podName || !session?.container) return null;
        const existing = sessions.find(
            s =>
                s.namespace === session.namespace &&
                s.podName === session.podName &&
                s.container === session.container
        );
        if (existing) {
            setActiveId(existing.id);
            return existing;
        }
        const newSession = { ...session, id: session.id || createSessionId() };
        setSessions(prev => [...prev, newSession]);
        setActiveId(newSession.id);
        return newSession;
    };

    const removeSession = (id) => {
        setSessions(prev => {
            const remaining = prev.filter(s => s.id !== id);
            if (activeId === id) {
                setActiveId(remaining[remaining.length - 1]?.id || null);
            }
            return remaining;
        });
    };

    const clearSessions = () => {
        setSessions([]);
        setActiveId(null);
    };

    const value = useMemo(
        () => ({
            sessions,
            activeId,
            setActiveId,
            addSession,
            removeSession,
            clearSessions,
        }),
        [sessions, activeId]
    );

    return (
        <TerminalDockContext.Provider value={value}>
            {children}
        </TerminalDockContext.Provider>
    );
};

export const useTerminalDock = () => useContext(TerminalDockContext);
