package runtime

import "context"

type Executor interface {
	Run(ctx context.Context, input map[string]any, config map[string]any) (map[string]any, error)
}
