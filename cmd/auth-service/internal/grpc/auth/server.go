package auth

import (
	"context"
	"errors"
	"immxrtalbeast/order_microservices/auth-service/internal/domain"
	"immxrtalbeast/order_microservices/auth-service/internal/services/auth"

	ssov2 "github.com/immxrtalbeast/order_protos/gen/go/auth"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type serverAPI struct {
	ssov2.UnimplementedAuthServer
	auth Auth
}

type Auth interface {
	Login(
		ctx context.Context,
		email string,
		password string,
	) (token string, err error)
	RegisterNewUser(
		ctx context.Context,
		email string,
		password string,
	) (userID uuid.UUID, err error)
	IsAdmin(ctx context.Context, userID uuid.UUID) (bool, error)
}

func Register(gRPCServer *grpc.Server, auth Auth) {
	ssov2.RegisterAuthServer(gRPCServer, &serverAPI{auth: auth})
}

func (s *serverAPI) Login(ctx context.Context, in *ssov2.LoginRequest) (*ssov2.LoginResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	token, err := s.auth.Login(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			return nil, status.Error(codes.InvalidArgument, "invalid email or password")
		}

		return nil, status.Error(codes.Internal, "failed to login")
	}

	return &ssov2.LoginResponse{Token: token}, nil
}

func (s *serverAPI) Register(ctx context.Context, in *ssov2.RegisterRequest) (*ssov2.RegisterResponse, error) {
	if in.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "email is required")
	}

	if in.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	uid, err := s.auth.RegisterNewUser(ctx, in.GetEmail(), in.GetPassword())
	if err != nil {
		if errors.Is(err, domain.ErrUserExists) {
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}

		return nil, status.Error(codes.Internal, "failed to register user")
	}

	return &ssov2.RegisterResponse{UserId: uid.String()}, nil
}

func (s *serverAPI) IsAdmin(ctx context.Context, in *ssov2.IsAdminRequest) (*ssov2.IsAdminResponse, error) {
	if in.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user ID is required")
	}

	userID, err := uuid.Parse(in.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user ID format")
	}

	isAdmin, err := s.auth.IsAdmin(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "failed to check admin role")
	}

	return &ssov2.IsAdminResponse{IsAdmin: isAdmin}, nil
}
