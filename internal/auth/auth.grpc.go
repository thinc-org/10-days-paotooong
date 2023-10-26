package auth

import (
	"context"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/auth/v1"
)

type authServiceImpl struct {
	v1.UnimplementedAuthServiceServer
	
	client *ent.Client
}

func NewService(client *ent.Client) v1.AuthServiceServer {
	return &authServiceImpl{ 
		v1.UnimplementedAuthServiceServer{},
		client,
	}
}

func (s *authServiceImpl) Login(ctx context.Context, request *v1.LoginRequest) (*v1.LoginResponse, error) {
	return nil, nil
}

func (s *authServiceImpl) Register(ctx context.Context, request *v1.RegisterRequest) (*v1.RegisterResponse, error) {
	_, err := s.client.User.
		Create().
		SetEmail(request.Email).
		SetHash("555").
		SetFirstName(request.FirstName).
		SetFamilyName(request.FamilyName).
		Save(ctx)
	
	if err != nil {
		return nil, err
	}
	
	return nil, nil
}

func (s *authServiceImpl) Me(ctx context.Context, request *v1.MeRequest) (*v1.MeResponse, error) {
	return nil, nil
}
