package waitprocess

import "context"

type Process interface {
	Run()
	Stop()
	SetContext(ctx context.Context)
}
