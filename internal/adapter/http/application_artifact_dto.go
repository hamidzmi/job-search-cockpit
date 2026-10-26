package httpadapter

import (
	"fmt"
	"time"

	"job-search-cockpit/internal/domain"
)

type ApplicationArtifactDTO struct {
	ID              string `json:"id"`
	OpportunityID   string `json:"opportunity_id"`
	Kind            string `json:"kind"`
	Title           string `json:"title"`
	ContentMarkdown string `json:"content_markdown"`
	HasAttachedFile bool   `json:"has_attached_file"`
	FileName        string `json:"file_name"`
	MimeType        string `json:"mime_type"`
	FileSizeBytes   int64  `json:"file_size_bytes"`
	DownloadURL     string `json:"download_url"`
	IsSubmitted     bool   `json:"is_submitted"`
	ModelUsed       string `json:"model_used"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type ApplicationArtifactUpsertRequest struct {
	Kind            string `json:"kind"`
	Title           string `json:"title"`
	ContentMarkdown string `json:"content_markdown"`
	IsSubmitted     bool   `json:"is_submitted"`
	ModelUsed       string `json:"model_used"`
}

type ArtifactDraftRequest struct {
	Kind            string `json:"kind"`
	Title           string `json:"title"`
	ExistingContent string `json:"existing_content"`
	Instructions    string `json:"instructions"`
	ModelOverride   string `json:"model_override"`
}

type ArtifactDraftDTO struct {
	Kind            string `json:"kind"`
	Title           string `json:"title"`
	ContentMarkdown string `json:"content_markdown"`
	ModelUsed       string `json:"model_used"`
}

func (r ApplicationArtifactUpsertRequest) toDomainInput() domain.ApplicationArtifactInput {
	return domain.ApplicationArtifactInput{
		Kind:            domain.ArtifactKind(r.Kind),
		Title:           r.Title,
		ContentMarkdown: r.ContentMarkdown,
		IsSubmitted:     r.IsSubmitted,
		ModelUsed:       r.ModelUsed,
	}
}

func (r ArtifactDraftRequest) toDomainInput() domain.ArtifactDraftRequestInput {
	return domain.ArtifactDraftRequestInput{
		Kind:            domain.ArtifactKind(r.Kind),
		Title:           r.Title,
		ExistingContent: r.ExistingContent,
		Instructions:    r.Instructions,
		ModelOverride:   r.ModelOverride,
	}
}

func toApplicationArtifactDTO(art domain.ApplicationArtifact) ApplicationArtifactDTO {
	downloadURL := ""
	if art.HasAttachedFile() {
		downloadURL = fmt.Sprintf("/api/opportunities/%s/artifacts/%s/file", art.OpportunityID, art.ID)
	}
	return ApplicationArtifactDTO{
		ID:              art.ID,
		OpportunityID:   art.OpportunityID,
		Kind:            string(art.Kind),
		Title:           art.Title,
		ContentMarkdown: art.ContentMarkdown,
		HasAttachedFile: art.HasAttachedFile(),
		FileName:        art.FileName,
		MimeType:        art.MimeType,
		FileSizeBytes:   art.FileSizeBytes,
		DownloadURL:     downloadURL,
		IsSubmitted:     art.IsSubmitted,
		ModelUsed:       art.ModelUsed,
		CreatedAt:       art.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:       art.UpdatedAt.UTC().Format(time.RFC3339),
	}
}
