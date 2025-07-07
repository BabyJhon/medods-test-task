package service

import (
	"context"
	"net"

	"github.com/BabyJhon/medods-test-task/internal/entity"
	"github.com/BabyJhon/medods-test-task/internal/repo"
	"github.com/gofrs/uuid"
)

type Auth interface {
	generateAccessToken(userID, sessionID uuid.UUID, userAgent string, clientIP net.IP) (string, error)
	generateRefreshToken() ([]byte, error)
	CreateTokens(ctx context.Context, guid uuid.UUID, userAgent string, clientIP net.IP) (string, string, error)
	Parsetoken(accessToken string) (*entity.Claimes, error)
	GetSession(ctx context.Context, token entity.Claimes) (entity.Session, error)
	RevokeToken(ctx context.Context, sessionID uuid.UUID) error
	RefreshTokens(ctx context.Context, accessToken entity.Claimes, base64RefreshToken string, userAgent string, IP net.IP) (string, string, error)
}

type Service struct {
	Auth
}

func NewService(repos *repo.Repository) *Service {
	return &Service{
		Auth: NewAuthService(repos),
	}
}
