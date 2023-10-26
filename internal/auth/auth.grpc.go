package auth

import (
	"context"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
	"github.com/thinc-org/10-days-paotooong/gen/ent/user"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
	"github.com/thinc-org/10-days-paotooong/internal/token"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authServiceImpl struct {
	v1.UnimplementedAuthServiceServer
	
	client *ent.Client
	tokenSvc token.TokenService
}

func NewService(client *ent.Client, tokenSvc token.TokenService) v1.AuthServiceServer {
	return &authServiceImpl{ 
		v1.UnimplementedAuthServiceServer{},
		client,
		tokenSvc,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, request *v1.LoginRequest) (*v1.LoginResponse, error) {
	user, err := s.client.User.Query().Where(
		user.Email(request.GetEmail()),
	).First(ctx)

	if err != nil {
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
			Ttl: int32(ttl),
		},
	}, nil
}

func (s *authServiceImpl) Register(ctx context.Context, request *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	pwdHash, err := hashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	user, err := s.client.User.
		Create().
		SetEmail(request.Email).
		SetHash(pwdHash).
		SetFirstName(request.FirstName).
		SetFamilyName(request.FamilyName).
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
			Ttl: int32(ttl),
		},
	}, nil
}

func (s *authServiceImpl) Me(ctx context.Context, request *v1.MeRequest) (*v1.MeResponse, error) {
	return nil, nil
}

func comparePasswordWithHash(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}

func hashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
    return string(bytes), err
}
