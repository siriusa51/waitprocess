package waitprocess

import "context"

type Process interface {
	Run() error
	Stop()
	SetContext(ctx context.Context)
}
