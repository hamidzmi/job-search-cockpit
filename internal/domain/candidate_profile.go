package domain

import (
	"fmt"
	"strings"
	"time"
)

type WorkHistoryEntry struct {
	Company       string
	Title         string
	StartDate     string
	EndDate       string
	Summary       string
	ImpactBullets []string
}

type CandidateProfile struct {
	FullName            string
	Headline            string
	Summary             string
	WorkHistory         []WorkHistoryEntry
	Skills              []string
	TargetRoleTitles    []string
	LocationPreferences []string
	RemotePreference    string
	CompensationFloor   int
	TargetCompensation  int
	Currency            string
	UpdatedAt           time.Time
}

type CandidateProfileInput struct {
	FullName            string
	Headline            string
	Summary             string
	WorkHistory         []WorkHistoryEntry
	Skills              []string
	TargetRoleTitles    []string
	LocationPreferences []string
	RemotePreference    string
	CompensationFloor   int
	TargetCompensation  int
	Currency            string
}

func DefaultCandidateProfile() CandidateProfile {
	return CandidateProfile{
		WorkHistory:         []WorkHistoryEntry{},
		Skills:              []string{},
		TargetRoleTitles:    []string{},
		LocationPreferences: []string{},
		Currency:            "USD",
	}
}

func NewCandidateProfile(in CandidateProfileInput, now time.Time) (CandidateProfile, error) {
	if in.CompensationFloor < 0 || in.TargetCompensation < 0 {
		return CandidateProfile{}, fmt.Errorf("%w: compensation cannot be negative", ErrValidation)
	}
	if in.CompensationFloor > 0 && in.TargetCompensation > 0 && in.TargetCompensation < in.CompensationFloor {
		return CandidateProfile{}, fmt.Errorf("%w: target compensation cannot be less than compensation floor", ErrValidation)
	}

	history := make([]WorkHistoryEntry, 0, len(in.WorkHistory))
	for _, entry := range in.WorkHistory {
		company := strings.TrimSpace(entry.Company)
		title := strings.TrimSpace(entry.Title)
		if company == "" || title == "" {
			return CandidateProfile{}, fmt.Errorf("%w: work history entry requires company and title", ErrValidation)
		}
		bullets := make([]string, 0, len(entry.ImpactBullets))
		for _, b := range entry.ImpactBullets {
			if trimmed := strings.TrimSpace(b); trimmed != "" {
				bullets = append(bullets, trimmed)
			}
		}
		history = append(history, WorkHistoryEntry{
			Company:       company,
			Title:         title,
			StartDate:     strings.TrimSpace(entry.StartDate),
			EndDate:       strings.TrimSpace(entry.EndDate),
			Summary:       strings.TrimSpace(entry.Summary),
			ImpactBullets: bullets,
		})
	}

	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "USD"
	}

	return CandidateProfile{
		FullName:            strings.TrimSpace(in.FullName),
		Headline:            strings.TrimSpace(in.Headline),
		Summary:             strings.TrimSpace(in.Summary),
		WorkHistory:         history,
		Skills:              normalizeUniqueStrings(in.Skills),
		TargetRoleTitles:    normalizeUniqueStrings(in.TargetRoleTitles),
		LocationPreferences: normalizeUniqueStrings(in.LocationPreferences),
		RemotePreference:    strings.TrimSpace(in.RemotePreference),
		CompensationFloor:   in.CompensationFloor,
		TargetCompensation:  in.TargetCompensation,
		Currency:            currency,
		UpdatedAt:           now.UTC(),
	}, nil
}

func normalizeUniqueStrings(items []string) []string {
	result := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}
