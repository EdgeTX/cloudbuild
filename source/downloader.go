package source

import (
	"context"
)

type Downloader interface {
	Download(ctx context.Context, repository string, commitID string) error
}
