import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { MemoryRouter } from 'react-router-dom';
import { KanbanBoard } from '../KanbanBoard';
import type { Opportunity } from '../../../api/types';

describe('KanbanBoard Drag and Drop & Minimalist Display', () => {
  const mockOpportunity: Opportunity = {
    id: 'opp-dnd-1',
    company_id: 'comp-1',
    company_name: 'Vercel',
    role_title: 'Staff Frontend Infrastructure Engineer',
    seniority: 'Staff',
    location: 'Remote',
    remote_policy: 'Remote',
    priority: 'High',
    stage: 'Discovered',
    outcome: 'Active',
    notes: 'Direct referral',
    job_posting: {
      source_url: '',
      raw_text: '',
      requirements: [],
      tech_stack_tags: ['React', 'TypeScript'],
      stated_min_comp: 0,
      stated_max_comp: 0,
      currency: '',
      has_content: true,
    },
    last_activity_at: '2026-10-25T10:00:00Z',
    created_at: '2026-10-25T10:00:00Z',
    updated_at: '2026-10-25T10:00:00Z',
  };

  it('allows dragging an opportunity card and dropping onto another stage column to advance stage', async () => {
    const handleMoveStage = vi.fn().mockResolvedValue(undefined);

    render(
      <MemoryRouter>
        <KanbanBoard
          opportunities={[mockOpportunity]}
          onMoveStage={handleMoveStage}
        />
      </MemoryRouter>
    );

    const card = screen.getByText('Staff Frontend Infrastructure Engineer').closest('article')!;
    expect(card).toBeInTheDocument();
    expect(card).toHaveAttribute('draggable', 'true');

    // Simulate drag start on card
    const dataTransfer = {
      setData: vi.fn(),
      getData: vi.fn().mockReturnValue('opp-dnd-1'),
      effectAllowed: 'none',
      dropEffect: 'none',
    };

    fireEvent.dragStart(card, { dataTransfer });
    expect(dataTransfer.setData).toHaveBeenCalledWith('text/plain', 'opp-dnd-1');

    // Find the 'Interviewing' column target
    const interviewingCol = screen.getByRole('region', { name: 'Interviewing Stage Column' });
    expect(interviewingCol).toBeInTheDocument();

    // Drag over and drop onto column
    fireEvent.dragOver(interviewingCol, { dataTransfer });
    fireEvent.drop(interviewingCol, { dataTransfer });

    expect(handleMoveStage).toHaveBeenCalledWith('opp-dnd-1', 'Interviewing');

    // Toast feedback appears
    expect(
      await screen.findByText(/Moved Staff Frontend Infrastructure Engineer to Interviewing/i)
    ).toBeInTheDocument();
  });

  it('allows changing outcome on an opportunity card and undoing the action', async () => {
    const handleMoveStage = vi.fn().mockResolvedValue(undefined);
    const handleChangeOutcome = vi.fn().mockResolvedValue(undefined);

    render(
      <MemoryRouter>
        <KanbanBoard
          opportunities={[mockOpportunity]}
          onMoveStage={handleMoveStage}
          onChangeOutcome={handleChangeOutcome}
        />
      </MemoryRouter>
    );

    const outcomeSelect = screen.getByLabelText('Set Staff Frontend Infrastructure Engineer Outcome');
    expect(outcomeSelect).toHaveValue('Active');

    fireEvent.change(outcomeSelect, { target: { value: 'Rejected' } });

    expect(handleChangeOutcome).toHaveBeenCalledWith('opp-dnd-1', 'Rejected');

    const toastText = await screen.findByText(
      /Marked Staff Frontend Infrastructure Engineer as Rejected/i
    );
    expect(toastText).toBeInTheDocument();

    const undoButton = screen.getByRole('button', { name: 'Undo' });
    expect(undoButton).toBeInTheDocument();

    fireEvent.click(undoButton);
    expect(handleChangeOutcome).toHaveBeenCalledWith(
      'opp-dnd-1',
      'Active',
      'Reverted to Active (Undo)'
    );
  });
});
