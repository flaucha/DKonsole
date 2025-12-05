import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, within } from '@testing-library/react';
import PodDetails from './PodDetails';
import { useAuth } from '../../context/AuthContext';
import { useTerminalDock } from '../../context/TerminalDockContext';

// Mocks
vi.mock('../../context/AuthContext');
vi.mock('../../context/TerminalDockContext');
vi.mock('../../utils/logger', () => ({
    logger: {
        error: vi.fn(),
    },
}));
vi.mock('../../utils/dateUtils', () => ({
    formatDateTime: (date) => `Formatted: ${date}`,
    formatDateTimeShort: (date) => `Short: ${date}`,
}));
vi.mock('./CommonDetails', () => ({
    DetailRow: ({ label, value }) => <div data-testid={`detail-row-${label}`}>{label}: {value}</div>,
    EditYamlButton: ({ onClick }) => <button onClick={onClick}>Edit YAML</button>,
}));
vi.mock('./LogViewerInline', () => ({
    default: () => <div data-testid="log-viewer">LogViewer</div>,
}));
vi.mock('./TerminalViewerInline', () => ({
    default: () => <div data-testid="terminal-viewer">TerminalViewer</div>,
}));
vi.mock('../PodMetrics', () => ({
    default: () => <div data-testid="pod-metrics">PodMetrics</div>,
}));

describe('PodDetails', () => {
    const mockAuthFetch = vi.fn();
    const mockAddSession = vi.fn();
    const mockOnEditYAML = vi.fn();

    const mockDetails = {
        node: 'node-1',
        ip: '10.0.0.1',
        restarts: 5,
        containers: ['container-1', 'container-2'],
        metrics: {
            cpu: '100m',
            memory: '256Mi',
        },
        containerStatuses: [
            {
                name: 'container-1',
                ready: true,
                state: 'Running',
                restartCount: 2,
                startedAt: '2023-01-01T00:00:00Z',
                image: 'image-1:latest',
            },
            {
                name: 'container-2',
                ready: false,
                state: 'Waiting',
                reason: 'CrashLoopBackOff',
                message: 'Back-off restarting failed container',
                restartCount: 3,
                image: 'image-2:latest',
            },
        ],
    };

    const mockPod = {
        name: 'test-pod',
        namespace: 'default',
    };

    beforeEach(() => {
        vi.mocked(useAuth).mockReturnValue({ authFetch: mockAuthFetch });
        vi.mocked(useTerminalDock).mockReturnValue({ addSession: mockAddSession });
        mockAuthFetch.mockReset();
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should render details tab by default', () => {
        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        expect(screen.getByText('Details')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('detail-row-Node')).toHaveTextContent('Node: node-1');
        expect(screen.getByTestId('detail-row-IP')).toHaveTextContent('IP: 10.0.0.1');
        expect(screen.getByTestId('detail-row-Restarts')).toHaveTextContent('Restarts: 5');
        // Containers might be rendered as joined string or array, check generic content
        expect(screen.getByText(/container-1/)).toBeInTheDocument();
        expect(screen.getByText('Edit YAML')).toBeInTheDocument();
    });

    it('should switch to logs tab', async () => {
        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        fireEvent.click(screen.getByText('Logs'));

        expect(screen.getByText('Logs')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('log-viewer')).toBeInTheDocument();

        // Should show container selector because we have multiple containers
        expect(screen.getByRole('combobox')).toBeInTheDocument();
        expect(screen.getByRole('combobox')).toHaveValue('container-1');
    });

    it('should switch to terminal tab', () => {
        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        fireEvent.click(screen.getByText('Terminal'));

        expect(screen.getByText('Terminal')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('terminal-viewer')).toBeInTheDocument();
        expect(screen.getByRole('combobox')).toBeInTheDocument();
    });

    it('should switch to metrics tab', () => {
        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        fireEvent.click(screen.getByText('Metrics'));

        expect(screen.getByText('Metrics')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('pod-metrics')).toBeInTheDocument();
    });

    it('should switch to events tab and fetch events', async () => {
        const mockEvents = [
            { type: 'Normal', reason: 'Started', message: 'Started container', count: 1, lastSeen: '2023-01-01T00:00:00Z' },
            { type: 'Warning', reason: 'Failed', message: 'Failed to pull image', count: 2, lastSeen: '2023-01-01T00:01:00Z' },
        ];

        mockAuthFetch.mockResolvedValue({
            ok: true,
            json: async () => mockEvents,
        });

        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        fireEvent.click(screen.getByText('Events'));

        expect(screen.getByText('Events')).toHaveClass('bg-gray-700');
        expect(screen.getByText('Loading events...')).toBeInTheDocument();

        await waitFor(() => {
            expect(screen.queryByText('Loading events...')).not.toBeInTheDocument();
        });

        expect(mockAuthFetch).toHaveBeenCalledWith(expect.stringContaining('/api/pods/events'));
        expect(screen.getByText('Started')).toBeInTheDocument();
        expect(screen.getByText('Failed')).toBeInTheDocument();
        expect(screen.getByText('x2')).toBeInTheDocument();
    });

    it('should handle event fetch error', async () => {
        mockAuthFetch.mockRejectedValue(new Error('Fetch error'));

        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        fireEvent.click(screen.getByText('Events'));

        await waitFor(() => {
            expect(screen.queryByText('Loading events...')).not.toBeInTheDocument();
        });

        // Should show empty state or logs error (we mocked logger)
        // The component sets empty events on error
        expect(screen.getByText('No events available')).toBeInTheDocument();
    });

    it('should render container status timeline', async () => {
        // Mock fetch to avoid crash when switching to events tab
        mockAuthFetch.mockResolvedValue({
            ok: true,
            json: async () => [],
        });

        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);

        // Use Events tab to see timeline? No, wait.
        // Looking at PodDetails.jsx structure:
        // The last tab is what? 
        // TabButton active={activeTab === 'events'} label="Events"
        // If activeTab is 'events' it shows events. 
        // Wait, where is the timeline rendered?
        // Line 238: {/* Container Status Timeline */}
        // It is inside the `activeTab === 'metrics' ? ... : ( ... )` else block of events?
        // No. 
        // Line 181: activeTab === 'metrics' ? ...
        // Line 185: : ( ... // This is the ELSE for metrics.
        // Is this else for Events??
        // Line 89: onClick={() => setActiveTab('events')}
        // So if activeTab is 'events', it goes to the ELSE block at 185?
        // Let's check logic:
        // activeTab === 'details' ? (...) :
        // activeTab === 'logs' ? (...) :
        // activeTab === 'terminal' ? (...) :
        // activeTab === 'metrics' ? (...) :
        // ( ... // THIS MUST BE EVENTS

        // So the "Events" tab actually renders BOTH "Pod Events" AND "Container Status Timeline".

        fireEvent.click(screen.getByText('Events'));

        // Check container statuses
        // container-1 is Running
        const c1 = screen.getByText('container-1');
        expect(c1).toBeInTheDocument();
        // Running status
        expect(screen.getByText('Running')).toBeInTheDocument();

        // container-2 is Waiting
        const c2 = screen.getByText('container-2');
        expect(c2).toBeInTheDocument();
        expect(screen.getByText('Waiting')).toBeInTheDocument();
        expect(screen.getByText('Reason:')).toBeInTheDocument();
        expect(screen.getByText('CrashLoopBackOff')).toBeInTheDocument();
    });

    it('should allow editing YAML', () => {
        render(<PodDetails details={mockDetails} onEditYAML={mockOnEditYAML} pod={mockPod} />);
        fireEvent.click(screen.getByText('Edit YAML'));
        expect(mockOnEditYAML).toHaveBeenCalled();
    });
});
