package context

import (
	"context"

	"github.com/iamtraining/gallery/models"
)

type privateKey string

const (
	userKey privateKey = "user"
)

func WithUser(ctx context.Context, user *models.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func GetUser(ctx context.Context) *models.User {
	if v := ctx.Value(userKey); v != nil {
		if user, ok := v.(*models.User); ok {
			return user
		}
	}

	return nil
}
