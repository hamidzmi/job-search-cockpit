package httpadapter

import (
	"time"

	"job-search-cockpit/internal/domain"
)

type UpcomingInterviewRoundDTO struct {
	ID                   string   `json:"id"`
	OpportunityID        string   `json:"opportunity_id"`
	CompanyName          string   `json:"company_name"`
	OpportunityRoleTitle string   `json:"opportunity_role_title"`
	RoundOrder           int      `json:"round_order"`
	RoundType            string   `json:"round_type"`
	Title                string   `json:"title"`
	ScheduledAt          string   `json:"scheduled_at"`
	DurationMinutes      int      `json:"duration_minutes"`
	Status               string   `json:"status"`
	PrepNotes            string   `json:"prep_notes"`
	ContactIDs           []string `json:"contact_ids"`
	StoryIDs             []string `json:"story_ids"`
	InterviewerNames     []string `json:"interviewer_names"`
	StoryTitles          []string `json:"story_titles"`
}

type MissingDebriefRoundDTO struct {
	ID                   string `json:"id"`
	OpportunityID        string `json:"opportunity_id"`
	CompanyName          string `json:"company_name"`
	OpportunityRoleTitle string `json:"opportunity_role_title"`
	RoundOrder           int    `json:"round_order"`
	RoundType            string `json:"round_type"`
	Title                string `json:"title"`
	ScheduledAt          string `json:"scheduled_at"`
	DurationMinutes      int    `json:"duration_minutes"`
	Status               string `json:"status"`
}

type FlightDeckActionItemDTO struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	OpportunityID        string `json:"opportunity_id,omitempty"`
	OpportunityRoleTitle string `json:"opportunity_role_title,omitempty"`
	ContactID            string `json:"contact_id,omitempty"`
	ContactName          string `json:"contact_name,omitempty"`
	CompanyName          string `json:"company_name,omitempty"`
	DueAt                string `json:"due_at"`
	Status               string `json:"status"`
	Notes                string `json:"notes,omitempty"`
	CreatedAt            string `json:"created_at"`
	UpdatedAt            string `json:"updated_at"`
	IsOverdue            bool   `json:"is_overdue"`
	IsDueToday           bool   `json:"is_due_today"`
}

type StaleOpportunityDTO struct {
	OpportunityID  string `json:"opportunity_id"`
	CompanyID      string `json:"company_id"`
	CompanyName    string `json:"company_name"`
	RoleTitle      string `json:"role_title"`
	Stage          string `json:"stage"`
	Outcome        string `json:"outcome"`
	Priority       string `json:"priority"`
	StaleReason    string `json:"stale_reason"`
	DaysStale      int    `json:"days_stale"`
	AnchorDate     string `json:"anchor_date"`
	LastActivityAt string `json:"last_activity_at"`
}

type FlightDeckDTO struct {
	UpcomingRounds     []UpcomingInterviewRoundDTO `json:"upcoming_rounds"`
	MissingDebriefs    []MissingDebriefRoundDTO    `json:"missing_debriefs"`
	ActionItems        []FlightDeckActionItemDTO   `json:"action_items"`
	StaleOpportunities []StaleOpportunityDTO       `json:"stale_opportunities"`
}

func toUpcomingRoundDTO(v domain.UpcomingInterviewRoundView) UpcomingInterviewRoundDTO {
	var schedStr string
	if !v.Round.ScheduledAt.IsZero() {
		schedStr = v.Round.ScheduledAt.UTC().Format(time.RFC3339Nano)
	}
	contactIDs := v.Round.ContactIDs
	if contactIDs == nil {
		contactIDs = []string{}
	}
	storyIDs := v.Round.StoryIDs
	if storyIDs == nil {
		storyIDs = []string{}
	}
	interviewerNames := v.InterviewerNames
	if interviewerNames == nil {
		interviewerNames = []string{}
	}
	storyTitles := v.StoryTitles
	if storyTitles == nil {
		storyTitles = []string{}
	}
	return UpcomingInterviewRoundDTO{
		ID:                   v.Round.ID,
		OpportunityID:        v.Round.OpportunityID,
		CompanyName:          v.CompanyName,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		RoundOrder:           v.Round.RoundOrder,
		RoundType:            string(v.Round.RoundType),
		Title:                v.Round.Title,
		ScheduledAt:          schedStr,
		DurationMinutes:      v.Round.DurationMinutes,
		Status:               string(v.Round.Status),
		PrepNotes:            v.Round.PrepNotes,
		ContactIDs:           contactIDs,
		StoryIDs:             storyIDs,
		InterviewerNames:     interviewerNames,
		StoryTitles:          storyTitles,
	}
}

func toMissingDebriefRoundDTO(v domain.MissingDebriefRoundView) MissingDebriefRoundDTO {
	var schedStr string
	if !v.Round.ScheduledAt.IsZero() {
		schedStr = v.Round.ScheduledAt.UTC().Format(time.RFC3339Nano)
	}
	return MissingDebriefRoundDTO{
		ID:                   v.Round.ID,
		OpportunityID:        v.Round.OpportunityID,
		CompanyName:          v.CompanyName,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		RoundOrder:           v.Round.RoundOrder,
		RoundType:            string(v.Round.RoundType),
		Title:                v.Round.Title,
		ScheduledAt:          schedStr,
		DurationMinutes:      v.Round.DurationMinutes,
		Status:               string(v.Round.Status),
	}
}

func toFlightDeckActionItemDTO(v domain.FlightDeckActionItemView) FlightDeckActionItemDTO {
	return FlightDeckActionItemDTO{
		ID:                   v.ActionItem.ID,
		Title:                v.ActionItem.Title,
		OpportunityID:        v.ActionItem.OpportunityID,
		OpportunityRoleTitle: v.OpportunityRoleTitle,
		ContactID:            v.ActionItem.ContactID,
		ContactName:          v.ContactName,
		CompanyName:          v.CompanyName,
		DueAt:                v.ActionItem.DueAt.UTC().Format(time.RFC3339Nano),
		Status:               string(v.ActionItem.Status),
		Notes:                v.ActionItem.Notes,
		CreatedAt:            v.ActionItem.CreatedAt.UTC().Format(time.RFC3339Nano),
		UpdatedAt:            v.ActionItem.UpdatedAt.UTC().Format(time.RFC3339Nano),
		IsOverdue:            v.IsOverdue,
		IsDueToday:           v.IsDueToday,
	}
}

func toStaleOpportunityDTO(v domain.StaleOpportunityView) StaleOpportunityDTO {
	var anchorStr string
	if !v.AnchorDate.IsZero() {
		anchorStr = v.AnchorDate.UTC().Format(time.RFC3339Nano)
	}
	var actStr string
	if !v.LastActivityAt.IsZero() {
		actStr = v.LastActivityAt.UTC().Format(time.RFC3339Nano)
	}
	return StaleOpportunityDTO{
		OpportunityID:  v.OpportunityID,
		CompanyID:      v.CompanyID,
		CompanyName:    v.CompanyName,
		RoleTitle:      v.RoleTitle,
		Stage:          string(v.Stage),
		Outcome:        string(v.Outcome),
		Priority:       string(v.Priority),
		StaleReason:    v.StaleReason,
		DaysStale:      v.DaysStale,
		AnchorDate:     anchorStr,
		LastActivityAt: actStr,
	}
}

func toFlightDeckDTO(fd domain.FlightDeck) FlightDeckDTO {
	upcoming := make([]UpcomingInterviewRoundDTO, 0, len(fd.UpcomingRounds))
	for _, u := range fd.UpcomingRounds {
		upcoming = append(upcoming, toUpcomingRoundDTO(u))
	}

	missing := make([]MissingDebriefRoundDTO, 0, len(fd.MissingDebriefs))
	for _, m := range fd.MissingDebriefs {
		missing = append(missing, toMissingDebriefRoundDTO(m))
	}

	actions := make([]FlightDeckActionItemDTO, 0, len(fd.ActionItems))
	for _, a := range fd.ActionItems {
		actions = append(actions, toFlightDeckActionItemDTO(a))
	}

	stale := make([]StaleOpportunityDTO, 0, len(fd.StaleOpportunities))
	for _, s := range fd.StaleOpportunities {
		stale = append(stale, toStaleOpportunityDTO(s))
	}

	return FlightDeckDTO{
		UpcomingRounds:     upcoming,
		MissingDebriefs:    missing,
		ActionItems:        actions,
		StaleOpportunities: stale,
	}
}
