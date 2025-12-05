import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import DeploymentDetails from './DeploymentDetails';
import { useAuth } from '../../context/AuthContext';

// Mocks
vi.mock('../../context/AuthContext');
vi.mock('../../utils/permissions', () => ({
    canEdit: vi.fn(),
    isAdmin: vi.fn(),
}));
vi.mock('./CommonDetails', () => ({
    DetailRow: ({ label, value, children }) => (
        <div data-testid={`detail-row-${label}`}>
            {label}: {Array.isArray(value) ? value.join(', ') : value} {children}
        </div>
    ),
    SmartImage: ({ image }) => <span>{image}</span>,
    EditYamlButton: ({ onClick }) => <button onClick={onClick}>Edit YAML</button>,
}));
vi.mock('./AssociatedPods', () => ({
    default: () => <div data-testid="associated-pods">AssociatedPods</div>,
}));

import { canEdit, isAdmin } from '../../utils/permissions';

describe('DeploymentDetails', () => {
    const mockUser = { role: 'user', permissions: {} };
    const mockOnScale = vi.fn();
    const mockOnEditYAML = vi.fn();

    const mockDetails = {
        replicas: 3,
        ready: 3,
        images: ['image:v1'],
        ports: [8080],
        pvcs: ['pvc-1'],
        podLabels: { app: 'test-app' },
    };

    const mockRes = {
        name: 'test-deploy',
        namespace: 'default',
    };

    beforeEach(() => {
        vi.mocked(useAuth).mockReturnValue({ user: mockUser });
        vi.mocked(isAdmin).mockReturnValue(false);
        vi.mocked(canEdit).mockReturnValue(false);
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    it('should render details tab by default', () => {
        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        expect(screen.getByText('Details')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('detail-row-Replicas')).toHaveTextContent('Replicas: 3 / 3');
        expect(screen.getByTestId('detail-row-Images')).toHaveTextContent('Images:'); // Custom render in mock
        expect(screen.getByTestId('detail-row-Ports')).toHaveTextContent('Ports: 8080');
        expect(screen.getByTestId('detail-row-PVCs')).toHaveTextContent('PVCs: pvc-1');
        expect(screen.getByTestId('detail-row-Labels')).toHaveTextContent('Labels: app=test-app');
        expect(screen.getByText('Edit YAML')).toBeInTheDocument();
    });

    it('should switch to Pod List tab', () => {
        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        fireEvent.click(screen.getByText('Pod List'));

        expect(screen.getByText('Pod List')).toHaveClass('bg-gray-700');
        expect(screen.getByTestId('associated-pods')).toBeInTheDocument();
    });

    it('should show scale buttons if user is admin', () => {
        vi.mocked(isAdmin).mockReturnValue(true);

        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        expect(screen.getByTitle('Scale down')).toBeInTheDocument();
        expect(screen.getByTitle('Scale up')).toBeInTheDocument();
    });

    it('should show scale buttons if user can edit', () => {
        vi.mocked(canEdit).mockReturnValue(true);

        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        expect(screen.getByTitle('Scale down')).toBeInTheDocument();
        expect(screen.getByTitle('Scale up')).toBeInTheDocument();
    });

    it('should NOT show scale buttons if user cannot edit', () => {
        vi.mocked(isAdmin).mockReturnValue(false);
        vi.mocked(canEdit).mockReturnValue(false);

        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        expect(screen.queryByTitle('Scale down')).not.toBeInTheDocument();
        expect(screen.queryByTitle('Scale up')).not.toBeInTheDocument();
    });

    it('should call onScale when buttons clicked', () => {
        vi.mocked(isAdmin).mockReturnValue(true);

        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={false}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        fireEvent.click(screen.getByTitle('Scale down'));
        expect(mockOnScale).toHaveBeenCalledWith(-1);

        fireEvent.click(screen.getByTitle('Scale up'));
        expect(mockOnScale).toHaveBeenCalledWith(1);
    });

    it('should disable scale buttons when scaling', () => {
        vi.mocked(isAdmin).mockReturnValue(true);

        render(
            <DeploymentDetails
                details={mockDetails}
                onScale={mockOnScale}
                scaling={true}
                res={mockRes}
                onEditYAML={mockOnEditYAML}
            />
        );

        expect(screen.getByTitle('Scale down')).toBeDisabled();
        expect(screen.getByTitle('Scale up')).toBeDisabled();
    });
});
