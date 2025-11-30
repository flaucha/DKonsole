import React, { createContext, useContext, useState, useEffect } from 'react';
import { useAuth } from './AuthContext';
import { logger } from '../utils/logger';

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
    const [menuAnimation, setMenuAnimation] = useState(localStorage.getItem('menuAnimation') || 'slide');
    const [menuAnimationSpeed, setMenuAnimationSpeed] = useState(localStorage.getItem('menuAnimationSpeed') || 'medium');
    const [itemsPerPage, setItemsPerPage] = useState(() => {
        const saved = localStorage.getItem('itemsPerPage');
        return saved ? parseInt(saved, 10) : 500;
    });

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

    useEffect(() => {
        localStorage.setItem('menuAnimation', menuAnimation);
    }, [menuAnimation]);

    useEffect(() => {
        localStorage.setItem('menuAnimationSpeed', menuAnimationSpeed);
    }, [menuAnimationSpeed]);

    useEffect(() => {
        localStorage.setItem('itemsPerPage', itemsPerPage.toString());
    }, [itemsPerPage]);

    const fetchClusters = async () => {
        try {
            // Multi-cluster support was removed, always use 'default' cluster
            setClusters(['default']);
            setCurrentCluster('default');
        } catch (error) {
            logger.error('Failed to initialize clusters:', error);
            // Fallback to default
            setClusters(['default']);
            setCurrentCluster('default');
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
            setBorderRadius,
            menuAnimation,
            setMenuAnimation,
            menuAnimationSpeed,
            setMenuAnimationSpeed,
            itemsPerPage,
            setItemsPerPage
        }}>
            {children}
        </SettingsContext.Provider>
    );
};
