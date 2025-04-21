package embedding

import "context"

type Embedding interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}
