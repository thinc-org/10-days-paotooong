package interceptor

import (
	"context"
	"strings"

	"github.com/thinc-org/10-days-paotooong/internal/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func NewAuthInterceptor(tokenSvc token.TokenService) *AuthInterceptor {
	return &AuthInterceptor{
		tokenSvc,
	}
}

type AuthInterceptor struct {
	tokenSvc token.TokenService
}

func (i *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		if len(md.Get("authorization")) <= 0 {
			return handler(ctx, req)
		}
		headerValue := md.Get("authorization")[0]
		if !strings.HasPrefix(headerValue, "Bearer ") {
			return nil, status.Error(codes.FailedPrecondition, "unsupported authorization header")
		}

		token := strings.TrimPrefix(headerValue, "Bearer ")

		uid, err := i.tokenSvc.Validate(token)
		if err == nil {
			return handler(context.WithValue(ctx, "uid", uid), req)
		}

		return handler(ctx, req)
	}
}
