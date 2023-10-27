package auth

import (
	"context"
	"net/mail"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
	"github.com/thinc-org/10-days-paotooong/gen/ent/user"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/wallet/v1"
	"github.com/thinc-org/10-days-paotooong/internal/token"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"github.com/bufbuild/protovalidate-go"
)

var _ v1.WalletServiceServer = &walletServiceImpl{}

type walletServiceImpl struct {
	v1.UnimplementedWalletServiceServer
	
	client *ent.Client
	tokenSvc token.TokenService
}

func NewService(client *ent.Client, tokenSvc token.TokenService) v1.WalletServiceServer {
	return &walletServiceImpl{ 
		v1.UnimplementedWalletServiceServer{},
		client,
		tokenSvc,
	}
}

func (s *walletServiceImpl) Pay(ctx context.Context, request *v1.PayRequest) (*v1.PayResponse, error) {
	return nil, nil
}

func (s *walletServiceImpl) Topup(ctx context.Context, request *v1.TopupRequest) (*v1.TopupResponse, error) {
	return nil, nil
}

