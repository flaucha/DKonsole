import React, { createContext, useContext, useState, useEffect } from 'react';
import { useAuth } from './AuthContext';

const SettingsContext = createContext();

export const useSettings = () => useContext(SettingsContext);

export const SettingsProvider = ({ children }) => {
    const { authFetch } = useAuth();
    const [clusters, setClusters] = useState([]);
    const [currentCluster, setCurrentCluster] = useState('default');
    const [theme, setTheme] = useState(localStorage.getItem('theme') || 'default');
    const [font, setFont] = useState(localStorage.getItem('font') || 'Inter');
    const [fontSize, setFontSize] = useState(localStorage.getItem('fontSize') || 'normal');
    const [borderRadius, setBorderRadius] = useState(localStorage.getItem('borderRadius') || 'md');

    useEffect(() => {
        fetchClusters();
    }, []);

    useEffect(() => {
        document.documentElement.setAttribute('data-theme', theme);
        localStorage.setItem('theme', theme);
    }, [theme]);

    useEffect(() => {
        document.documentElement.style.setProperty('--font-family', font);
        localStorage.setItem('font', font);
    }, [font]);

    useEffect(() => {
        document.documentElement.setAttribute('data-font-size', fontSize);
        localStorage.setItem('fontSize', fontSize);
    }, [fontSize]);

    useEffect(() => {
        document.documentElement.setAttribute('data-border-radius', borderRadius);
        localStorage.setItem('borderRadius', borderRadius);
    }, [borderRadius]);

    const fetchClusters = async () => {
        try {
            const res = await authFetch('/api/clusters');
            if (res.ok) {
                const data = await res.json();
                setClusters(data);
                if (!data.includes(currentCluster)) {
                    setCurrentCluster(data[0] || 'default');
                }
            }
        } catch (error) {
            console.error('Failed to fetch clusters:', error);
        }
    };

    return (
        <SettingsContext.Provider value={{
            clusters,
            currentCluster,
            setCurrentCluster,
            theme,
            setTheme,
            font,
            setFont,
            fontSize,
            setFontSize,
            borderRadius,
            setBorderRadius
        }}>
            {children}
        </SettingsContext.Provider>
    );
};
