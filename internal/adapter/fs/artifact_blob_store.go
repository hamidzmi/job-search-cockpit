package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"job-search-cockpit/internal/domain"
)

type LocalArtifactBlobStore struct {
	baseDir string
}

func NewLocalArtifactBlobStore(baseDir string) *LocalArtifactBlobStore {
	if strings.TrimSpace(baseDir) == "" {
		baseDir = filepath.Join(".", "data", "artifacts")
	}
	return &LocalArtifactBlobStore{
		baseDir: filepath.Clean(baseDir),
	}
}

func (s *LocalArtifactBlobStore) SaveArtifactBlob(_ context.Context, opportunityID, artifactID, _ string, content []byte) (string, error) {
	oppKey := sanitizeSegment(opportunityID)
	artKey := sanitizeSegment(artifactID)
	if oppKey == "" || artKey == "" {
		return "", fmt.Errorf("%w: invalid artifact storage key", domain.ErrValidation)
	}
	if len(content) == 0 {
		return "", fmt.Errorf("%w: artifact blob content is empty", domain.ErrValidation)
	}

	storageKey := filepath.ToSlash(filepath.Join(oppKey, artKey+".pdf"))
	fullPath, err := s.resolveSafePath(storageKey)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return "", fmt.Errorf("create artifact directory: %w", err)
	}
	if err := os.WriteFile(fullPath, content, 0o644); err != nil {
		return "", fmt.Errorf("write artifact blob: %w", err)
	}

	return storageKey, nil
}

func (s *LocalArtifactBlobStore) ReadArtifactBlob(_ context.Context, storageKey string) ([]byte, error) {
	fullPath, err := s.resolveSafePath(storageKey)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: artifact file %s", domain.ErrNotFound, storageKey)
		}
		return nil, fmt.Errorf("read artifact blob: %w", err)
	}
	return data, nil
}

func (s *LocalArtifactBlobStore) DeleteArtifactBlob(_ context.Context, storageKey string) error {
	if strings.TrimSpace(storageKey) == "" {
		return nil
	}
	fullPath, err := s.resolveSafePath(storageKey)
	if err != nil {
		return err
	}
	if err := os.Remove(fullPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete artifact blob: %w", err)
	}
	return nil
}

func (s *LocalArtifactBlobStore) resolveSafePath(storageKey string) (string, error) {
	trimmed := strings.TrimSpace(storageKey)
	if trimmed == "" {
		return "", fmt.Errorf("%w: empty artifact storage key", domain.ErrValidation)
	}
	cleanedRel := filepath.Clean(filepath.FromSlash(trimmed))
	if filepath.IsAbs(cleanedRel) || strings.HasPrefix(cleanedRel, "..") {
		return "", fmt.Errorf("%w: invalid artifact storage path", domain.ErrValidation)
	}
	absBase, err := filepath.Abs(s.baseDir)
	if err != nil {
		return "", err
	}
	absTarget, err := filepath.Abs(filepath.Join(absBase, cleanedRel))
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return "", fmt.Errorf("%w: artifact path escapes base directory", domain.ErrValidation)
	}
	return absTarget, nil
}

func sanitizeSegment(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	var b strings.Builder
	for _, r := range trimmed {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
