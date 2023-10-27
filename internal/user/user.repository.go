package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/thinc-org/10-days-paotooong/gen/ent"
	"github.com/thinc-org/10-days-paotooong/gen/ent/user"
	"github.com/thinc-org/10-days-paotooong/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ UserRepository = &userRepositoryImpl{}

type UserRepository interface {
	FindUserById(ctx context.Context, uid string) (*ent.User, error)
	InferUserFromContext(ctx context.Context) (*ent.User, error)
}

type userRepositoryImpl struct {
	client *ent.Client
}

func (r *userRepositoryImpl) FindUserById(ctx context.Context, uid string) (*ent.User, error) {
	uuid, err := uuid.Parse(uid)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid uid")
	}

	user, err := r.client.User.Query().Where(user.ID(uuid)).First(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepositoryImpl) InferUserFromContext(ctx context.Context) (*ent.User, error) {
	uid, err := utils.InferUidFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Unauthorized")
	}

	return r.FindUserById(ctx, uid)
}
