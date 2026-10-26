package fs_test

import (
	"bytes"
	"context"
	"errors"
	"testing"

	fsadapter "job-search-cockpit/internal/adapter/fs"
	"job-search-cockpit/internal/domain"
)

func TestLocalArtifactBlobStore(t *testing.T) {
	store := fsadapter.NewLocalArtifactBlobStore(t.TempDir())
	ctx := context.Background()
	pdf := []byte("%PDF-1.7\nhello world\n%%EOF")

	key, err := store.SaveArtifactBlob(ctx, "opp-123", "art-456", "resume.pdf", pdf)
	if err != nil {
		t.Fatalf("SaveArtifactBlob failed: %v", err)
	}
	if key != "opp-123/art-456.pdf" {
		t.Fatalf("expected storage key opp-123/art-456.pdf, got %q", key)
	}

	got, err := store.ReadArtifactBlob(ctx, key)
	if err != nil {
		t.Fatalf("ReadArtifactBlob failed: %v", err)
	}
	if !bytes.Equal(got, pdf) {
		t.Fatalf("expected %q, got %q", pdf, got)
	}

	if _, err := store.ReadArtifactBlob(ctx, "../outside.pdf"); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("expected ErrValidation for path traversal, got %v", err)
	}

	if err := store.DeleteArtifactBlob(ctx, key); err != nil {
		t.Fatalf("DeleteArtifactBlob failed: %v", err)
	}
	if _, err := store.ReadArtifactBlob(ctx, key); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
