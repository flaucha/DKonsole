import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import logoFullDark from '../assets/logo-full-dark.png';
import logoFullLight from '../assets/logo-full-light.png';
import { logger } from '../utils/logger';

export const useSetupLogic = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [confirmPassword, setConfirmPassword] = useState('');
    const [jwtSecret, setJwtSecret] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState('');
    const [loading, setLoading] = useState(false);
    const [reloading, setReloading] = useState(false);
    const [setupCompleted, setSetupCompleted] = useState(false);
    const [checkingStatus, setCheckingStatus] = useState(true);
    const navigate = useNavigate();

    // Get current theme and determine default logo immediately
    const currentTheme = localStorage.getItem('theme') || 'default';
    const isLightTheme = currentTheme === 'light' || currentTheme === 'cream';
    const defaultLogoSrc = isLightTheme ? logoFullLight : logoFullDark;
    const [logoSrc, setLogoSrc] = useState(defaultLogoSrc);

    useEffect(() => {
        // Check if setup is already completed
        const checkSetupStatus = async () => {
            try {
                const res = await fetch('/api/setup/status');
                if (res.ok) {
                    const data = await res.json();
                    if (!data.setupRequired) {
                        // Setup is already completed
                        setSetupCompleted(true);
                    }
                }
            } catch (err) {
                logger.error('Failed to check setup status:', err);
            } finally {
                setCheckingStatus(false);
            }
        };

        checkSetupStatus();

        // Try to load custom logo from API (no auth required for logo endpoint)
        const logoType = isLightTheme ? 'light' : 'normal';
        const timestamp = Date.now();
        fetch(`/api/logo?type=${logoType}&t=${timestamp}`)
            .then(res => {
                if (res.ok && res.status === 200) {
                    setLogoSrc(`/api/logo?type=${logoType}&t=${timestamp}`);
                }
            })
            .catch(() => {
                // On error, keep default logo (already set)
            });
    }, [isLightTheme]);

    const handleLogoError = () => {
        // If logo fails to load, fallback to theme-appropriate default logo
        if (logoSrc !== defaultLogoSrc) {
            setLogoSrc(defaultLogoSrc);
        }
    };

    const generateJWTSecret = () => {
        // Generate a secure random JWT secret (44 characters base64)
        const array = new Uint8Array(32);
        crypto.getRandomValues(array);
        const base64 = btoa(String.fromCharCode(...array));
        setJwtSecret(base64);
    };

    const validateForm = () => {
        if (!username.trim()) {
            setError('Username is required');
            return false;
        }

        if (!password) {
            setError('Password is required');
            return false;
        }

        if (password.length < 8) {
            setError('Password must be at least 8 characters long');
            return false;
        }

        if (password !== confirmPassword) {
            setError('Passwords do not match');
            return false;
        }

        if (jwtSecret && jwtSecret.length < 32) {
            setError('JWT secret must be at least 32 characters long');
            return false;
        }

        return true;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        setSuccess('');

        if (!validateForm()) {
            return;
        }

        setLoading(true);

        try {
            const response = await fetch('/api/setup/complete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    username,
                    password,
                    jwtSecret: jwtSecret || undefined, // Send undefined if empty to trigger auto-generation
                }),
            });

            const data = await response.json();

            if (response.ok) {
                // Hide form and show reloading screen
                setReloading(true);
                setLoading(false);

                // Wait 5 seconds then navigate to login
                setTimeout(() => {
                    navigate('/login');
                }, 5000);
            } else {
                setError(data.message || data.error || 'Failed to complete setup');
                setLoading(false);
            }
        } catch (err) {
            logger.error('Setup failed:', err);
            setError('An error occurred while completing setup');
            setLoading(false);
        }
    };

    return {
        // State
        username, setUsername,
        password, setPassword,
        confirmPassword, setConfirmPassword,
        jwtSecret, setJwtSecret,
        error, success, loading, reloading,
        setupCompleted, checkingStatus,
        logoSrc, handleLogoError,

        // Handlers
        generateJWTSecret,
        handleSubmit,
        navigate
    };
};
