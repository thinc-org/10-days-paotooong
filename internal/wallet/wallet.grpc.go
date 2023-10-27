package auth

import (
	"context"

	"github.com/thinc-org/10-days-paotooong/gen/ent"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/wallet/v1"
	user_repo "github.com/thinc-org/10-days-paotooong/internal/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ v1.WalletServiceServer = &walletServiceImpl{}

type walletServiceImpl struct {
	v1.UnimplementedWalletServiceServer
	
	client *ent.Client
	userRepo user_repo.UserRepository
}

func NewService(client *ent.Client, userRepo user_repo.UserRepository) v1.WalletServiceServer {
	return &walletServiceImpl{ 
		v1.UnimplementedWalletServiceServer{},
		client,
		userRepo,
	}
}

func (s *walletServiceImpl) Pay(ctx context.Context, request *v1.PayRequest) (*v1.PayResponse, error) {
	payer, err := s.userRepo.InferUserFromContext(ctx)

	if err != nil {
		return nil, err
	}

	if float32(payer.Money) < request.GetAmount() {
		return nil, status.Error(codes.FailedPrecondition, "not enough money")
	}
	
	receiverId := request.ReceiverId
	receiver, err := s.userRepo.FindUserById(ctx, receiverId)

	receiver.Update().SetMoney(float64(request.Amount) + float64(request.GetAmount())).Save(ctx)
	payer.Update().SetMoney(payer.Money - float64(request.GetAmount())).Save(ctx)
	transaction, err := s.client.Transaction.Create().AddPayer(payer).AddReceiver(receiver).Save(ctx)

	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.PayResponse{
		Transaction: &v1.Transaction{
			TransactionId: transaction.ID.String(),
			Payer: &v1.UserTransaction{
				Id: payer.ID.String(),
				FirstName: payer.FirstName,
				FamilyName: payer.FamilyName,
			},
			Receiver: &v1.UserTransaction{
				Id: receiverId,
				FirstName: receiver.FirstName,
				FamilyName: receiver.FamilyName,
			},
			Amount: request.GetAmount(),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transaction.CreatedAt.Unix(),
			},
		},
	}, nil
}

func (s *walletServiceImpl) Topup(ctx context.Context, request *v1.TopupRequest) (*v1.TopupResponse, error) {
	return nil, nil
}

