package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

const singletonCandidateProfileID = "default"

type workHistoryRecord struct {
	Company       string   `json:"company"`
	Title         string   `json:"title"`
	StartDate     string   `json:"start_date"`
	EndDate       string   `json:"end_date"`
	Summary       string   `json:"summary"`
	ImpactBullets []string `json:"impact_bullets"`
}

func (s *Store) SaveCandidateProfile(ctx context.Context, p domain.CandidateProfile) error {
	records := make([]workHistoryRecord, 0, len(p.WorkHistory))
	for _, entry := range p.WorkHistory {
		records = append(records, workHistoryRecord{
			Company:       entry.Company,
			Title:         entry.Title,
			StartDate:     entry.StartDate,
			EndDate:       entry.EndDate,
			Summary:       entry.Summary,
			ImpactBullets: nonNilSlice(entry.ImpactBullets),
		})
	}
	historyJSON, err := json.Marshal(records)
	if err != nil {
		return err
	}
	skillsJSON, err := json.Marshal(nonNilSlice(p.Skills))
	if err != nil {
		return err
	}
	titlesJSON, err := json.Marshal(nonNilSlice(p.TargetRoleTitles))
	if err != nil {
		return err
	}
	locationsJSON, err := json.Marshal(nonNilSlice(p.LocationPreferences))
	if err != nil {
		return err
	}

	_, err = s.exec(ctx).ExecContext(ctx, `
		INSERT INTO candidate_profile (
			id, full_name, headline, summary, work_history_json, skills_json,
			target_role_titles_json, location_preferences_json, remote_preference,
			compensation_floor, target_compensation, currency, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			full_name = excluded.full_name,
			headline = excluded.headline,
			summary = excluded.summary,
			work_history_json = excluded.work_history_json,
			skills_json = excluded.skills_json,
			target_role_titles_json = excluded.target_role_titles_json,
			location_preferences_json = excluded.location_preferences_json,
			remote_preference = excluded.remote_preference,
			compensation_floor = excluded.compensation_floor,
			target_compensation = excluded.target_compensation,
			currency = excluded.currency,
			updated_at = excluded.updated_at
	`,
		singletonCandidateProfileID,
		p.FullName,
		p.Headline,
		p.Summary,
		string(historyJSON),
		string(skillsJSON),
		string(titlesJSON),
		string(locationsJSON),
		p.RemotePreference,
		p.CompensationFloor,
		p.TargetCompensation,
		p.Currency,
		p.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	return err
}

func (s *Store) GetCandidateProfile(ctx context.Context) (domain.CandidateProfile, error) {
	var p domain.CandidateProfile
	var historyJSON, skillsJSON, titlesJSON, locationsJSON, updatedStr string

	err := s.exec(ctx).QueryRowContext(ctx, `
		SELECT
			full_name, headline, summary, work_history_json, skills_json,
			target_role_titles_json, location_preferences_json, remote_preference,
			compensation_floor, target_compensation, currency, updated_at
		FROM candidate_profile
		WHERE id = ?
	`, singletonCandidateProfileID).Scan(
		&p.FullName,
		&p.Headline,
		&p.Summary,
		&historyJSON,
		&skillsJSON,
		&titlesJSON,
		&locationsJSON,
		&p.RemotePreference,
		&p.CompensationFloor,
		&p.TargetCompensation,
		&p.Currency,
		&updatedStr,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.CandidateProfile{}, fmt.Errorf("%w: candidate profile", domain.ErrNotFound)
	}
	if err != nil {
		return domain.CandidateProfile{}, err
	}

	var records []workHistoryRecord
	_ = json.Unmarshal([]byte(historyJSON), &records)
	p.WorkHistory = make([]domain.WorkHistoryEntry, 0, len(records))
	for _, r := range records {
		bullets := r.ImpactBullets
		if bullets == nil {
			bullets = []string{}
		}
		p.WorkHistory = append(p.WorkHistory, domain.WorkHistoryEntry{
			Company:       r.Company,
			Title:         r.Title,
			StartDate:     r.StartDate,
			EndDate:       r.EndDate,
			Summary:       r.Summary,
			ImpactBullets: bullets,
		})
	}
	_ = json.Unmarshal([]byte(skillsJSON), &p.Skills)
	_ = json.Unmarshal([]byte(titlesJSON), &p.TargetRoleTitles)
	_ = json.Unmarshal([]byte(locationsJSON), &p.LocationPreferences)
	if p.Skills == nil {
		p.Skills = []string{}
	}
	if p.TargetRoleTitles == nil {
		p.TargetRoleTitles = []string{}
	}
	if p.LocationPreferences == nil {
		p.LocationPreferences = []string{}
	}
	p.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedStr)
	return p, nil
}
