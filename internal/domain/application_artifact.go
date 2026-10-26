package domain

import (
	"fmt"
	"strings"
	"time"
)

type ArtifactKind string

const (
	ArtifactKindResume      ArtifactKind = "resume"
	ArtifactKindCoverLetter ArtifactKind = "cover_letter"
	ArtifactKindScreeningQA ArtifactKind = "screening_qa"
	ArtifactKindOther       ArtifactKind = "other"
)

func AllArtifactKinds() []ArtifactKind {
	return []ArtifactKind{
		ArtifactKindResume,
		ArtifactKindCoverLetter,
		ArtifactKindScreeningQA,
		ArtifactKindOther,
	}
}

func (k ArtifactKind) IsValid() bool {
	switch k {
	case ArtifactKindResume, ArtifactKindCoverLetter, ArtifactKindScreeningQA, ArtifactKindOther:
		return true
	default:
		return false
	}
}

type ApplicationArtifact struct {
	ID              string
	OpportunityID   string
	Kind            ArtifactKind
	Title           string
	ContentMarkdown string
	FilePath        string
	FileName        string
	MimeType        string
	FileSizeBytes   int64
	IsSubmitted     bool
	ModelUsed       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ApplicationArtifactInput struct {
	Kind            ArtifactKind
	Title           string
	ContentMarkdown string
	IsSubmitted     bool
	ModelUsed       string
}

type ArtifactDraftRequestInput struct {
	Kind            ArtifactKind
	Title           string
	ExistingContent string
	Instructions    string
	ModelOverride   string
}

type AssistantDraftArtifactInput struct {
	Opportunity     OpportunityView
	Profile         CandidateProfile
	Stories         []Story
	FitAssessment   *FitAssessment
	Kind            ArtifactKind
	Title           string
	ExistingContent string
	Instructions    string
	ModelOverride   string
}

type AssistantArtifactDraft struct {
	Kind            ArtifactKind
	Title           string
	ContentMarkdown string
	ModelUsed       string
}

type ArtifactFilePayload struct {
	FileName    string
	MimeType    string
	Content     []byte
}

func NewApplicationArtifact(id, opportunityID string, in ApplicationArtifactInput, now time.Time) (ApplicationArtifact, error) {
	trimmedID := strings.TrimSpace(id)
	if trimmedID == "" {
		return ApplicationArtifact{}, fmt.Errorf("%w: artifact id is required", ErrValidation)
	}
	trimmedOppID := strings.TrimSpace(opportunityID)
	if trimmedOppID == "" {
		return ApplicationArtifact{}, fmt.Errorf("%w: opportunity id is required", ErrValidation)
	}

	ts := now.UTC()
	art := ApplicationArtifact{
		ID:            trimmedID,
		OpportunityID: trimmedOppID,
		CreatedAt:     ts,
	}
	if err := art.Update(in, ts); err != nil {
		return ApplicationArtifact{}, err
	}
	return art, nil
}

func (a *ApplicationArtifact) Update(in ApplicationArtifactInput, now time.Time) error {
	kind := ArtifactKind(strings.TrimSpace(string(in.Kind)))
	if !kind.IsValid() {
		return fmt.Errorf("%w: invalid application artifact kind %q", ErrValidation, in.Kind)
	}
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return fmt.Errorf("%w: application artifact title is required", ErrValidation)
	}

	a.Kind = kind
	a.Title = title
	a.ContentMarkdown = strings.TrimSpace(in.ContentMarkdown)
	a.IsSubmitted = in.IsSubmitted
	if trimmedModel := strings.TrimSpace(in.ModelUsed); trimmedModel != "" {
		a.ModelUsed = trimmedModel
	}
	a.UpdatedAt = now.UTC()
	return nil
}

func (a ApplicationArtifact) HasAttachedFile() bool {
	return strings.TrimSpace(a.FilePath) != ""
}

func (a *ApplicationArtifact) AttachFile(filePath, fileName, mimeType string, fileSizeBytes int64, now time.Time) error {
	trimmedPath := strings.TrimSpace(filePath)
	if trimmedPath == "" {
		return fmt.Errorf("%w: file path is required", ErrValidation)
	}
	trimmedName := strings.TrimSpace(fileName)
	if trimmedName == "" {
		trimmedName = "artifact.pdf"
	}
	normalizedMime := strings.ToLower(strings.TrimSpace(mimeType))
	if normalizedMime != "application/pdf" {
		return fmt.Errorf("%w: only application/pdf files are supported", ErrValidation)
	}
	if fileSizeBytes <= 0 {
		return fmt.Errorf("%w: uploaded PDF file cannot be empty", ErrValidation)
	}

	a.FilePath = trimmedPath
	a.FileName = trimmedName
	a.MimeType = "application/pdf"
	a.FileSizeBytes = fileSizeBytes
	a.UpdatedAt = now.UTC()
	return nil
}
