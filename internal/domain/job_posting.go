package domain

import (
	"fmt"
	"strings"
)

type JobPosting struct {
	SourceURL     string
	RawText       string
	Requirements  []string
	TechStackTags []string
	StatedMinComp int
	StatedMaxComp int
	Currency      string
}

func (jp JobPosting) HasContent() bool {
	return strings.TrimSpace(jp.SourceURL) != "" ||
		strings.TrimSpace(jp.RawText) != "" ||
		len(jp.Requirements) > 0 ||
		len(jp.TechStackTags) > 0 ||
		jp.StatedMinComp > 0 ||
		jp.StatedMaxComp > 0
}

func normalizeJobPosting(jp JobPosting) (JobPosting, error) {
	if jp.StatedMinComp < 0 || jp.StatedMaxComp < 0 {
		return JobPosting{}, fmt.Errorf("%w: stated compensation cannot be negative", ErrValidation)
	}
	if jp.StatedMinComp > 0 && jp.StatedMaxComp > 0 && jp.StatedMinComp > jp.StatedMaxComp {
		return JobPosting{}, fmt.Errorf("%w: stated min compensation cannot exceed stated max compensation", ErrValidation)
	}
	reqs := make([]string, 0, len(jp.Requirements))
	for _, r := range jp.Requirements {
		if trimmed := strings.TrimSpace(r); trimmed != "" {
			reqs = append(reqs, trimmed)
		}
	}
	tags := make([]string, 0, len(jp.TechStackTags))
	for _, t := range jp.TechStackTags {
		if trimmed := strings.TrimSpace(t); trimmed != "" {
			tags = append(tags, trimmed)
		}
	}
	currency := strings.TrimSpace(jp.Currency)
	if currency == "" && (jp.StatedMinComp > 0 || jp.StatedMaxComp > 0) {
		currency = "USD"
	}
	return JobPosting{
		SourceURL:     strings.TrimSpace(jp.SourceURL),
		RawText:       strings.TrimSpace(jp.RawText),
		Requirements:  reqs,
		TechStackTags: tags,
		StatedMinComp: jp.StatedMinComp,
		StatedMaxComp: jp.StatedMaxComp,
		Currency:      currency,
	}, nil
}
