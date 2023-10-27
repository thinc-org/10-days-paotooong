package utils

import (
	"context"
	"errors"
)

func InferUidFromContext(ctx context.Context) (string, error) {
	uid, ok := ctx.Value("uid").(string)
	if !ok {
		return "", errors.New("no uid in context")
	}

	return uid, nil
}
