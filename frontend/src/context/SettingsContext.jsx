import React, { createContext, useContext, useState, useEffect } from 'react';

const SettingsContext = createContext();

export const useSettings = () => useContext(SettingsContext);

export const SettingsProvider = ({ children }) => {
    const [clusters, setClusters] = useState([]);
    const [currentCluster, setCurrentCluster] = useState('default');
    const [theme, setTheme] = useState(localStorage.getItem('theme') || 'default');
    const [font, setFont] = useState(localStorage.getItem('font') || 'Inter');

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

    const fetchClusters = async () => {
        try {
            const token = localStorage.getItem('token');
            const headers = token ? { 'Authorization': `Bearer ${token}` } : {};

            const res = await fetch('/api/clusters', { headers });
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

    const addCluster = async (config) => {
        try {
            const token = localStorage.getItem('token');
            const headers = {
                'Content-Type': 'application/json',
                ...(token && { 'Authorization': `Bearer ${token}` })
            };

            const res = await fetch('/api/clusters', {
                method: 'POST',
                headers,
                body: JSON.stringify(config),
            });
            if (res.ok) {
                await fetchClusters();
                return true;
            } else {
                const text = await res.text();
                throw new Error(text);
            }
        } catch (error) {
            console.error('Failed to add cluster:', error);
            throw error;
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
            addCluster
        }}>
            {children}
        </SettingsContext.Provider>
    );
};
