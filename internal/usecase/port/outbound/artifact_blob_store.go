package outbound

import "context"

type ArtifactBlobStore interface {
	SaveArtifactBlob(ctx context.Context, opportunityID, artifactID, fileName string, content []byte) (string, error)
	ReadArtifactBlob(ctx context.Context, storageKey string) ([]byte, error)
	DeleteArtifactBlob(ctx context.Context, storageKey string) error
}
