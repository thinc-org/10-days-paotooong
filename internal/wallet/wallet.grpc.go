package wallet

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/thinc-org/10-days-paotooong/gen/ent"
	"github.com/thinc-org/10-days-paotooong/gen/ent/pleasepay"
	v1 "github.com/thinc-org/10-days-paotooong/gen/proto/wallet/v1"
	user_repo "github.com/thinc-org/10-days-paotooong/internal/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ v1.WalletServiceServer = &walletServiceImpl{}

type walletServiceImpl struct {
	v1.UnimplementedWalletServiceServer

	client   *ent.Client
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
	if payer.ID.String() == receiverId {
		return nil, status.Error(codes.FailedPrecondition, "you cannot pay to yourself")
	}
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
				Id:         payer.ID.String(),
				FirstName:  payer.FirstName,
				FamilyName: payer.FamilyName,
			},
			Receiver: &v1.UserTransaction{
				Id:         receiverId,
				FirstName:  receiver.FirstName,
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
	if user.LastTopup != nil && time.Since(*user.LastTopup) < time.Duration(10*time.Minute) {
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
				Id:         user.ID.String(),
				FirstName:  user.FirstName,
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

func (s *walletServiceImpl) CreatePleasePay(ctx context.Context, request *v1.CreatePleasePayRequest) (*v1.CreatePleasePayResponse, error) {
	user, err := s.userRepo.InferUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	pleasePay, err := s.client.PleasePay.Create().SetReceiverID(user.ID).SetAmount(float64(request.GetAmount())).Save(ctx)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.CreatePleasePayResponse{
		PleasePay: &v1.PleasePay{
			Id:         pleasePay.ID.String(),
			State:      mapPleasePayState(pleasePay.State),
			ReceiverId: user.ID.String(),
			Amount:     float32(pleasePay.Amount),
		},
	}, nil
}

func (s *walletServiceImpl) GetPleasePay(ctx context.Context, request *v1.GetPleasePayRequest) (*v1.GetPleasePayResponse, error) {
	userEnt, err := s.userRepo.InferUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ppId, err := uuid.Parse(request.GetId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid uuid")
	}

	pleasePay, err := s.client.PleasePay.Query().Where(
		pleasepay.ID(ppId),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "please pay not found")
		}
		return nil, err
	}

	receiver, err := pleasePay.QueryReceiver().First(ctx)
	if receiver.ID != userEnt.ID {
		return nil, status.Error(codes.PermissionDenied, "this is not your please pay")
	}

	response := &v1.GetPleasePayResponse{
		PleasePay: &v1.PleasePay{
			Id:         pleasePay.ID.String(),
			State:      mapPleasePayState(pleasePay.State),
			ReceiverId: receiver.ID.String(),
			Amount:     float32(pleasePay.Amount),
		},
	}

	t, err := pleasePay.QueryTransaction().First(ctx)
	if err != nil && !ent.IsNotFound(err) {
		return nil, err
	}

	if t != nil {
		payer, err := t.QueryPayer().First(ctx)
		if err != nil {
			log.Printf("cannot find user %v in transaction %v", t.PayerID, t.ID)
			return nil, status.Error(codes.Internal, "internal server error")
		}

		receiver, err := t.QueryReceiver().First(ctx)
		if err != nil {
			log.Printf("cannot find user %v in transaction %v", t.ReceiverID, t.ID)
			return nil, status.Error(codes.Internal, "internal server error")
		}

		response.PleasePay.Transaction = &v1.Transaction{
			TransactionId: t.ID.String(),
			Payer: &v1.UserTransaction{
				Id:         payer.ID.String(),
				FirstName:  payer.FirstName,
				FamilyName: payer.FamilyName,
			},
			Receiver: &v1.UserTransaction{
				Id:         receiver.ID.String(),
				FirstName:  receiver.FirstName,
				FamilyName: receiver.FamilyName,
			},
			Amount: float32(t.Amount),
			CreatedAt: &timestamppb.Timestamp{
				Seconds: t.CreatedAt.Unix(),
			},
			Type: v1.TransactionType_TRANSACTION_TYPE_PAY,
		}
	}

	return response, nil
}

func (s *walletServiceImpl) PayPleasePay(ctx context.Context, request *v1.PayPleasePayRequest) (*v1.PayPleasePayResponse, error) {
	payer, err := s.userRepo.InferUserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ppId, err := uuid.Parse(request.GetId())
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, "invalid uuid")
	}

	pleasePay, err := s.client.PleasePay.Query().Where(
		pleasepay.ID(ppId),
	).First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "please pay not found")
		}
		return nil, err
	}

	receiver, err := pleasePay.QueryReceiver().First(ctx)
	if err != nil {
		return nil, status.Error(codes.NotFound, "receiver not found")
	}

	if payer.ID == receiver.ID {
		return nil, status.Error(codes.FailedPrecondition, "you cannot pay to yourself")
	}

	if float32(payer.Money) < float32(pleasePay.Amount) {
		return nil, status.Error(codes.FailedPrecondition, "not enough money")
	}

	amount := pleasePay.Amount

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

	pleasePayAfter, err := tx.PleasePay.UpdateOneID(pleasePay.ID).SetState("PAID").Save(ctx)
	if err != nil {
		tx.Rollback()
		return nil, status.Error(codes.Internal, "internal server error")
	}

	err = tx.Commit()
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &v1.PayPleasePayResponse{
		PleasePay: &v1.PleasePay{
			Id:         pleasePayAfter.ID.String(),
			State:      mapPleasePayState(pleasePayAfter.State),
			ReceiverId: receiver.ID.String(),
			Amount:     float32(pleasePay.Amount),
			Transaction: &v1.Transaction{
				TransactionId: transaction.ID.String(),
				Payer: &v1.UserTransaction{
					Id:         payer.ID.String(),
					FirstName:  payer.FirstName,
					FamilyName: payer.FamilyName,
				},
				Receiver: &v1.UserTransaction{
					Id:         receiver.ID.String(),
					FirstName:  receiver.FirstName,
					FamilyName: receiver.FamilyName,
				},
				Amount: float32(transaction.Amount),
				CreatedAt: &timestamppb.Timestamp{
					Seconds: transaction.CreatedAt.Unix(),
				},
				Type: v1.TransactionType_TRANSACTION_TYPE_PAY,
			},
		},
	}, nil
}

func mapPleasePayState(state string) v1.PleasePayState {
	switch state {
	case "PENDING":
		return v1.PleasePayState_PLEASE_PAY_STATE_PENDING
	case "PAID":
		return v1.PleasePayState_PLEASE_PAY_STATE_PAID
	default:
		return v1.PleasePayState_PLEASE_PAY_STATE_UNSPECIFIED
	}
}
