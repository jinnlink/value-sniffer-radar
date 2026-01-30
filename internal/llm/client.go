package llm

import "context"

type Client interface {
	Name() string
	Complete(ctx context.Context, prompt string) (string, error)
}
