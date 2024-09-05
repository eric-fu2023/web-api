package context

import "context"

const DefaultContextKey = "description"

func AppendCtx(ctx context.Context, key string, value string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	w, _ := ctx.Value(DefaultContextKey).(string)
	ctx = context.WithValue(ctx, key, w+value)

	return ctx
}
