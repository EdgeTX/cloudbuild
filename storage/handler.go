package storage

import (
	"context"
	"net/url"
)

type Handler interface {
	Upload(context context.Context, data []byte, fileName string) (*url.URL, error)
}
