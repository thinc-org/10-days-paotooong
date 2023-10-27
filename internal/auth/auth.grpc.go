package auth

import (
	"context"
	"net/mail"

	"github.com/bufbuild/protovalidate-go"
	"github.com/thinc-org/10-days-paotooong/gen/ent"
	"github.com/thinc-org/10-days-paotooong/gen/ent/user"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
	user_v1 "github.com/thinc-org/10-days-paotooong/gen/proto/user/v1"
	"github.com/thinc-org/10-days-paotooong/internal/token"
	user_repo "github.com/thinc-org/10-days-paotooong/internal/user"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ v1.AuthServiceServer = &authServiceImpl{}

type authServiceImpl struct {
	v1.UnimplementedAuthServiceServer

	client   *ent.Client
	tokenSvc token.TokenService
	userRepo user_repo.UserRepository
}

func NewService(client *ent.Client, tokenSvc token.TokenService, userRepo user_repo.UserRepository) v1.AuthServiceServer {
	return &authServiceImpl{
		v1.UnimplementedAuthServiceServer{},
		client,
		tokenSvc,
		userRepo,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, request *v1.LoginRequest) (*v1.LoginResponse, error) {
	user, err := s.client.User.Query().Where(
		user.Email(request.GetEmail()),
	).First(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, err
	}

	if !comparePasswordWithHash(user.Hash, request.Password) {
		return nil, status.Error(codes.Unauthenticated, "incorrect password")
	}

	uId := user.ID
	token := s.tokenSvc.CreateToken(uId.String())
	ttl := s.tokenSvc.TTL()

	return &v1.LoginResponse{
		Token: &v1.AuthToken{
			AccessToken: token,
			Ttl:         int32(ttl),
		},
	}, nil
}

func (s *authServiceImpl) Register(ctx context.Context, request *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	pwdHash, err := hashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	v, err := protovalidate.New()
	if err != nil {
		return nil, status.Error(codes.Internal, "unable to init validator")
	}

	if err = v.Validate(request); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "validation failed: %v", err)
	}

	// validate email
	_, err = mail.ParseAddress(request.Email)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid email")
	}

	user, err := s.client.User.
		Create().
		SetEmail(request.Email).
		SetHash(pwdHash).
		SetFirstName(request.FirstName).
		SetFamilyName(request.FamilyName).
		SetMoney(10000).
		Save(ctx)

	if err != nil {
		return nil, err
	}

	uId := user.ID
	token := s.tokenSvc.CreateToken(uId.String())
	ttl := s.tokenSvc.TTL()

	return &v1.RegisterResponse{
		Token: &v1.AuthToken{
			AccessToken: token,
			Ttl:         int32(ttl),
		},
	}, nil
}

func (s *authServiceImpl) Me(ctx context.Context, request *v1.MeRequest) (*v1.MeResponse, error) {
	user, err := s.userRepo.InferUserFromContext(ctx)

	if err != nil {
		return nil, err
	}

	return &v1.MeResponse{
		User: &user_v1.User{
			Id:         user.ID.String(),
			Email:      user.Email,
			Money:      int32(user.Money),
			FirstName:  user.FirstName,
			FamilyName: user.FamilyName,
		},
	}, nil
}

func comparePasswordWithHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}
