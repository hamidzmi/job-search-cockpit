package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"job-search-cockpit/internal/domain"
)

type debriefQuestionRecord struct {
	Question            string `json:"question"`
	Category            string `json:"category"`
	AnsweredWithStoryID string `json:"answered_with_story_id"`
	Notes               string `json:"notes"`
}

type debriefRecord struct {
	QuestionsAsked  []debriefQuestionRecord `json:"questions_asked"`
	WentWell        string                  `json:"went_well"`
	ToImprove       string                  `json:"to_improve"`
	Signals         string                  `json:"signals"`
	ConfidenceScore int                     `json:"confidence_score"`
	FollowUpNotes   string                  `json:"follow_up_notes"`
	UpdatedAt       string                  `json:"updated_at"`
}

func (s *Store) SaveInterviewRound(ctx context.Context, r domain.InterviewRound) error {
	contactsJSON, err := json.Marshal(nonNilSlice(r.ContactIDs))
	if err != nil {
		return err
	}
	storiesJSON, err := json.Marshal(nonNilSlice(r.StoryIDs))
	if err != nil {
		return err
	}

	var debriefJSONStr string
	if r.Debrief != nil {
		qRecords := make([]debriefQuestionRecord, 0, len(r.Debrief.QuestionsAsked))
		for _, q := range r.Debrief.QuestionsAsked {
			qRecords = append(qRecords, debriefQuestionRecord{
				Question:            q.Question,
				Category:            q.Category,
				AnsweredWithStoryID: q.AnsweredWithStoryID,
				Notes:               q.Notes,
			})
		}
		rec := debriefRecord{
			QuestionsAsked:  qRecords,
			WentWell:        r.Debrief.WentWell,
			ToImprove:       r.Debrief.ToImprove,
			Signals:         r.Debrief.Signals,
			ConfidenceScore: r.Debrief.ConfidenceScore,
			FollowUpNotes:   r.Debrief.FollowUpNotes,
			UpdatedAt:       r.Debrief.UpdatedAt.UTC().Format(time.RFC3339Nano),
		}
		b, err := json.Marshal(rec)
		if err != nil {
			return err
		}
		debriefJSONStr = string(b)
	}

	var scheduledAtStr string
	if !r.ScheduledAt.IsZero() {
		scheduledAtStr = r.ScheduledAt.UTC().Format(time.RFC3339Nano)
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO interview_rounds (
			id, opportunity_id, round_order, round_type, title,
			scheduled_at, duration_minutes, status, prep_notes,
			contact_ids_json, story_ids_json, debrief_json,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			round_order = excluded.round_order,
			round_type = excluded.round_type,
			title = excluded.title,
			scheduled_at = excluded.scheduled_at,
			duration_minutes = excluded.duration_minutes,
			status = excluded.status,
			prep_notes = excluded.prep_notes,
			contact_ids_json = excluded.contact_ids_json,
			story_ids_json = excluded.story_ids_json,
			debrief_json = excluded.debrief_json,
			updated_at = excluded.updated_at
	`,
		r.ID,
		r.OpportunityID,
		r.RoundOrder,
		string(r.RoundType),
		r.Title,
		scheduledAtStr,
		r.DurationMinutes,
		string(r.Status),
		r.PrepNotes,
		string(contactsJSON),
		string(storiesJSON),
		debriefJSONStr,
		r.CreatedAt.UTC().Format(time.RFC3339Nano),
		r.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetInterviewRound(ctx context.Context, id string) (domain.InterviewRound, error) {
	row := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			id, opportunity_id, round_order, round_type, title,
			scheduled_at, duration_minutes, status, prep_notes,
			contact_ids_json, story_ids_json, debrief_json,
			created_at, updated_at
		FROM interview_rounds
		WHERE id = ?
	`, id)
	round, err := scanInterviewRound(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.InterviewRound{}, fmt.Errorf("%w: interview round %s", domain.ErrNotFound, id)
	}
	return round, err
}

func (s *Store) ListInterviewRoundsByOpportunity(ctx context.Context, opportunityID string) ([]domain.InterviewRound, error) {
	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			id, opportunity_id, round_order, round_type, title,
			scheduled_at, duration_minutes, status, prep_notes,
			contact_ids_json, story_ids_json, debrief_json,
			created_at, updated_at
		FROM interview_rounds
		WHERE opportunity_id = ?
		ORDER BY round_order ASC, created_at ASC, id ASC
	`, opportunityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []domain.InterviewRound{}
	for rows.Next() {
		r, err := scanInterviewRound(rows.Scan)
		if err != nil {
			return nil, err
		}
		result = append(result, r)
	}
	return result, rows.Err()
}

func (s *Store) DeleteInterviewRound(ctx context.Context, id string) error {
	res, err := s.exec(ctx).ExecContext(ctx, `DELETE FROM interview_rounds WHERE id = ?`, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("%w: interview round %s", domain.ErrNotFound, id)
	}
	return nil
}

func (s *Store) ListQuestionBankEntries(ctx context.Context, filter domain.QuestionBankFilter) ([]domain.QuestionBankEntryView, error) {
	stories, err := s.ListStories(ctx, domain.StoryFilter{})
	if err != nil {
		return nil, err
	}
	storyTitles := make(map[string]string, len(stories))
	for _, st := range stories {
		storyTitles[st.ID] = st.Title
	}

	rows, err := s.exec(ctx).QueryContext(ctx, `
		SELECT
			ir.id, ir.opportunity_id, ir.round_order, ir.round_type, ir.title,
			ir.debrief_json, ir.updated_at,
			o.role_title, o.company_id, c.name
		FROM interview_rounds ir
		JOIN opportunities o ON o.id = ir.opportunity_id
		JOIN companies c ON c.id = o.company_id
		WHERE ir.debrief_json != ''
		ORDER BY ir.updated_at DESC, ir.round_order ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []domain.QuestionBankEntryView{}
	for rows.Next() {
		var roundID, oppID, roundTypeStr, roundTitle, debriefJSON, updatedStr string
		var oppRoleTitle, companyID, companyName string
		var roundOrder int

		if err := rows.Scan(
			&roundID,
			&oppID,
			&roundOrder,
			&roundTypeStr,
			&roundTitle,
			&debriefJSON,
			&updatedStr,
			&oppRoleTitle,
			&companyID,
			&companyName,
		); err != nil {
			return nil, err
		}

		var rec debriefRecord
		if err := json.Unmarshal([]byte(debriefJSON), &rec); err != nil {
			continue
		}
		recordedAt, err := time.Parse(time.RFC3339Nano, rec.UpdatedAt)
		if err != nil {
			recordedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
		}

		for _, q := range rec.QuestionsAsked {
			entry := domain.QuestionBankEntryView{
				RoundID:                roundID,
				OpportunityID:          oppID,
				OpportunityRoleTitle:   oppRoleTitle,
				CompanyID:              companyID,
				CompanyName:            companyName,
				RoundOrder:             roundOrder,
				RoundType:              domain.InterviewRoundType(roundTypeStr),
				RoundTitle:             roundTitle,
				Question:               q.Question,
				Category:               q.Category,
				AnsweredWithStoryID:    q.AnsweredWithStoryID,
				AnsweredWithStoryTitle: storyTitles[q.AnsweredWithStoryID],
				Notes:                  q.Notes,
				RecordedAt:             recordedAt,
			}
			if matchesQuestionBankFilter(entry, filter) {
				entries = append(entries, entry)
			}
		}
	}
	return entries, rows.Err()
}

func matchesQuestionBankFilter(entry domain.QuestionBankEntryView, f domain.QuestionBankFilter) bool {
	if wantCat := strings.TrimSpace(f.Category); wantCat != "" {
		if !strings.EqualFold(strings.TrimSpace(entry.Category), wantCat) {
			return false
		}
	}
	if wantCompany := strings.TrimSpace(f.CompanyID); wantCompany != "" {
		if entry.CompanyID != wantCompany {
			return false
		}
	}
	if wantType := strings.TrimSpace(string(f.RoundType)); wantType != "" {
		if !strings.EqualFold(strings.TrimSpace(string(entry.RoundType)), wantType) {
			return false
		}
	}
	if wantStory := strings.TrimSpace(f.StoryID); wantStory != "" {
		if entry.AnsweredWithStoryID != wantStory {
			return false
		}
	}
	if q := strings.ToLower(strings.TrimSpace(f.Query)); q != "" {
		haystack := strings.ToLower(strings.Join([]string{
			entry.Question,
			entry.Category,
			entry.Notes,
			entry.CompanyName,
			entry.OpportunityRoleTitle,
			string(entry.RoundType),
			entry.RoundTitle,
			entry.AnsweredWithStoryTitle,
		}, " "))
		if !strings.Contains(haystack, q) {
			return false
		}
	}
	return true
}

func scanInterviewRound(scan func(dest ...any) error) (domain.InterviewRound, error) {
	var r domain.InterviewRound
	var roundTypeStr, statusStr, scheduledStr, contactsJSON, storiesJSON, debriefJSON, createdStr, updatedStr string

	err := scan(
		&r.ID,
		&r.OpportunityID,
		&r.RoundOrder,
		&roundTypeStr,
		&r.Title,
		&scheduledStr,
		&r.DurationMinutes,
		&statusStr,
		&r.PrepNotes,
		&contactsJSON,
		&storiesJSON,
		&debriefJSON,
		&createdStr,
		&updatedStr,
	)
	if err != nil {
		return domain.InterviewRound{}, err
	}
	r.RoundType = domain.InterviewRoundType(roundTypeStr)
	r.Status = domain.InterviewRoundStatus(statusStr)
	if scheduledStr != "" {
		r.ScheduledAt, _ = time.Parse(time.RFC3339Nano, scheduledStr)
	}
	_ = json.Unmarshal([]byte(contactsJSON), &r.ContactIDs)
	_ = json.Unmarshal([]byte(storiesJSON), &r.StoryIDs)
	if r.ContactIDs == nil {
		r.ContactIDs = []string{}
	}
	if r.StoryIDs == nil {
		r.StoryIDs = []string{}
	}

	if debriefJSON != "" {
		var rec debriefRecord
		if err := json.Unmarshal([]byte(debriefJSON), &rec); err == nil {
			qs := make([]domain.DebriefQuestion, 0, len(rec.QuestionsAsked))
			for _, q := range rec.QuestionsAsked {
				qs = append(qs, domain.DebriefQuestion{
					Question:            q.Question,
					Category:            q.Category,
					AnsweredWithStoryID: q.AnsweredWithStoryID,
					Notes:               q.Notes,
				})
			}
			debriefUpdated, _ := time.Parse(time.RFC3339Nano, rec.UpdatedAt)
			r.Debrief = &domain.Debrief{
				QuestionsAsked:  qs,
				WentWell:        rec.WentWell,
				ToImprove:       rec.ToImprove,
				Signals:         rec.Signals,
				ConfidenceScore: rec.ConfidenceScore,
				FollowUpNotes:   rec.FollowUpNotes,
				UpdatedAt:       debriefUpdated,
			}
		}
	}

	r.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdStr)
	r.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return r, nil
}
