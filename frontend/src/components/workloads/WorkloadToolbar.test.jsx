import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { vi } from 'vitest';
import WorkloadToolbar from './WorkloadToolbar';

describe('WorkloadToolbar', () => {
    const defaultProps = {
        kind: 'Pod',
        filter: '',
        setFilter: vi.fn(),
        isSearchFocused: false,
        setIsSearchFocused: vi.fn(),
        refetch: vi.fn(),
        resourcesCount: 10,
        menuOpen: null,
        setMenuOpen: vi.fn(),
        orderedDataColumns: [
            { id: 'col1', label: 'Column 1' },
            { id: 'col2', label: 'Column 2' }
        ],
        hidden: [],
        toggleVisibility: vi.fn(),
        resetOrder: vi.fn(),
        onAdd: vi.fn()
    };

    it('renders filter input and count', () => {
        render(<WorkloadToolbar {...defaultProps} />);

        expect(screen.getByPlaceholderText('Filter...')).toBeInTheDocument();
        expect(screen.getByText('10 items')).toBeInTheDocument();
    });

    it('calls setFilter on input change', () => {
        render(<WorkloadToolbar {...defaultProps} />);

        const input = screen.getByPlaceholderText('Filter...');
        fireEvent.change(input, { target: { value: 'test' } });

        expect(defaultProps.setFilter).toHaveBeenCalledWith('test');
    });

    it('toggles column menu', () => {
        render(<WorkloadToolbar {...defaultProps} />);

        const columnsBtn = screen.getByTitle('Manage Columns');
        fireEvent.click(columnsBtn);

        expect(defaultProps.setMenuOpen).toHaveBeenCalledWith('columns');
    });

    it('renders column menu content when open', () => {
        render(<WorkloadToolbar {...defaultProps} menuOpen="columns" />);

        expect(screen.getByText('Visible Columns')).toBeInTheDocument();
        expect(screen.getByText('Column 1')).toBeInTheDocument();
        expect(screen.getByText('Column 2')).toBeInTheDocument();
    });

    it('toggles column visibility', () => {
        render(<WorkloadToolbar {...defaultProps} menuOpen="columns" />);

        const checkbox = screen.getAllByRole('checkbox')[0];
        fireEvent.click(checkbox);

        expect(defaultProps.toggleVisibility).toHaveBeenCalledWith('col1');
    });

    it('calls refetch on refresh click', () => {
        render(<WorkloadToolbar {...defaultProps} />);

        const refreshBtn = screen.getByTitle('Refresh');
        fireEvent.click(refreshBtn);

        expect(defaultProps.refetch).toHaveBeenCalled();
    });

    it('calls onAdd on add click', () => {
        render(<WorkloadToolbar {...defaultProps} />);

        const addBtn = screen.getByTitle('Add Resource');
        fireEvent.click(addBtn);

        expect(defaultProps.onAdd).toHaveBeenCalled();
    });
});
