import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import Settings from './Settings';
import { useSettings } from '../context/SettingsContext';

// Mock Dependencies
vi.mock('../context/SettingsContext');
vi.mock('./settings/ClustersSettings', () => ({ default: () => <div data-testid="clusters-settings">Clusters Settings Component</div> }));
vi.mock('./settings/AppearanceSettings', () => ({ default: () => <div data-testid="appearance-settings">Appearance Settings Component</div> }));
vi.mock('./settings/GeneralSettings', () => ({ default: () => <div data-testid="general-settings">General Settings Component</div> }));
vi.mock('./settings/LDAPSettings', () => ({ default: () => <div data-testid="ldap-settings">LDAP Settings Component</div> }));
vi.mock('./settings/AboutSettings', () => ({ default: () => <div data-testid="about-settings">About Settings Component</div> }));

describe('Settings Component', () => {
    const mockSetTheme = vi.fn();
    const mockSetFont = vi.fn();
    const mockSetFontSize = vi.fn();
    const mockSetBorderRadius = vi.fn();
    const mockSetMenuAnimationSpeed = vi.fn();

    beforeEach(() => {
        vi.clearAllMocks();
        vi.mocked(useSettings).mockReturnValue({
            setTheme: mockSetTheme,
            setFont: mockSetFont,
            setFontSize: mockSetFontSize,
            setBorderRadius: mockSetBorderRadius,
            setMenuAnimationSpeed: mockSetMenuAnimationSpeed
        });
    });

    it('should render Clusters tab by default', () => {
        render(<Settings />);

        expect(screen.getByText('Settings')).toBeInTheDocument();

        // Check active tab styling
        const clustersTab = screen.getByText('Clusters').closest('button');
        expect(clustersTab).toHaveClass('border-blue-500', 'text-blue-400');

        // Check content
        expect(screen.getByTestId('clusters-settings')).toBeInTheDocument();
        expect(screen.queryByTestId('appearance-settings')).not.toBeInTheDocument();
    });

    it('should switch to Appearance tab and content', () => {
        render(<Settings />);

        const appearanceTab = screen.getByText('Appearance');
        fireEvent.click(appearanceTab);

        // Check active tab styling
        expect(appearanceTab.closest('button')).toHaveClass('border-blue-500', 'text-blue-400');

        // Check content
        expect(screen.queryByTestId('clusters-settings')).not.toBeInTheDocument();
        expect(screen.getByTestId('appearance-settings')).toBeInTheDocument();
    });

    it('should show Reset Defaults button only on Appearance tab', () => {
        render(<Settings />);

        // Initially on Clusters, button should not be visible (or hidden)
        expect(screen.queryByText('Reset Defaults')).not.toBeInTheDocument();

        // Switch to Appearance
        fireEvent.click(screen.getByText('Appearance'));
        expect(screen.getByText('Reset Defaults')).toBeInTheDocument();

        // Switch to General
        fireEvent.click(screen.getByText('General'));
        expect(screen.queryByText('Reset Defaults')).not.toBeInTheDocument();
    });

    it('should call reset functions when Reset Defaults is clicked', () => {
        render(<Settings />);

        // Switch to Appearance
        fireEvent.click(screen.getByText('Appearance'));

        const resetButton = screen.getByText('Reset Defaults');
        fireEvent.click(resetButton);

        expect(mockSetTheme).toHaveBeenCalledWith('default');
        expect(mockSetFont).toHaveBeenCalledWith('Inter');
        expect(mockSetFontSize).toHaveBeenCalledWith('normal');
        expect(mockSetBorderRadius).toHaveBeenCalledWith('md');
        expect(mockSetMenuAnimationSpeed).toHaveBeenCalledWith('medium');
    });

    it('should switch to General tab', () => {
        render(<Settings />);
        fireEvent.click(screen.getByText('General'));
        expect(screen.getByTestId('general-settings')).toBeInTheDocument();
    });

    it('should switch to LDAP tab', () => {
        render(<Settings />);
        fireEvent.click(screen.getByText('LDAP'));
        expect(screen.getByTestId('ldap-settings')).toBeInTheDocument();
    });

    it('should switch to About tab', () => {
        render(<Settings />);
        fireEvent.click(screen.getByText('About'));
        expect(screen.getByTestId('about-settings')).toBeInTheDocument();
    });
});
