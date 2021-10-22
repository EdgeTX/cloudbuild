package firmware

import (
	"context"
)

type Builder interface {
	PullImage(ctx context.Context, buildContainer string) error
	Build(ctx context.Context, buildContainer string, flags []BuildFlag) ([]byte, error)
}
