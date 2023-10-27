package wallet

import (
	"context"
	"time"

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
	if err != nil {
		return nil, status.Error(codes.NotFound, "receiver not found")
	}

	amount := request.GetAmount()

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	_, err = tx.User.UpdateOneID(receiver.ID).SetMoney(receiver.Money + float64(amount)).Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	_, err = tx.User.UpdateOneID(payer.ID).SetMoney(payer.Money - float64(amount)).Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	transaction, err := tx.Transaction.Create().SetPayer(payer).SetReceiver(receiver).SetAmount(float64(amount)).SetType("pay").Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	err = tx.Commit()
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
			Type: v1.TransactionType_TRANSACTION_TYPE_PAY,
		},
	}, nil
}

func (s *walletServiceImpl) Topup(ctx context.Context, request *v1.TopupRequest) (*v1.TopupResponse, error) {
	user, err := s.userRepo.InferUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	amount := 100.0
	if user.LastTopup != nil && time.Since(*user.LastTopup) < time.Duration(10 * time.Minute) {
		return nil, status.Error(codes.FailedPrecondition, "last topup is within 10 minutes")
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	_, err = tx.User.UpdateOneID(user.ID).SetMoney(user.Money + amount).SetLastTopup(time.Now()).Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	transaction, err := tx.Transaction.Create().SetReceiver(user).SetAmount(amount).SetType("topup").Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.TopupResponse{
		Transaction: &v1.Transaction{
			TransactionId: transaction.ID.String(),
			Receiver: &v1.UserTransaction{
				Id: user.ID.String(),
				FirstName: user.FirstName,
				FamilyName: user.FamilyName,
			},
			Amount: float32(amount),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: transaction.CreatedAt.Unix(),
			},
			Type: v1.TransactionType_TRANSACTION_TYPE_TOPUP,
		},
	}, nil
}

