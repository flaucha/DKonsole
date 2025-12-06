import React from 'react';
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import AboutSettings from './AboutSettings';

describe('AboutSettings', () => {
    it('should render DKonsole title and description', () => {
        render(<AboutSettings />);

        expect(screen.getByText('DKonsole')).toBeInTheDocument();
        expect(screen.getByText(/A modern, lightweight Kubernetes dashboard/)).toBeInTheDocument();
    });

    it('should display version information', () => {
        render(<AboutSettings />);

        // Version might be mocked or environment dependent, looking for the label
        expect(screen.getByText('Version:')).toBeInTheDocument();
    });

    it('should display feature list', () => {
        render(<AboutSettings />);

        expect(screen.getByText(/Resource Management:/)).toBeInTheDocument();
        expect(screen.getByText(/Prometheus Integration:/)).toBeInTheDocument();
        expect(screen.getByText(/Live Logs:/)).toBeInTheDocument();
        expect(screen.getByText(/Terminal Access:/)).toBeInTheDocument();
    });

    it('should render external links', () => {
        render(<AboutSettings />);

        expect(screen.getByText('GitHub').closest('a')).toHaveAttribute('href', 'https://github.com/flaucha/DKonsole');
        expect(screen.getByText('Email').closest('a')).toHaveAttribute('href', 'mailto:flaucha@gmail.com');
    });
});
