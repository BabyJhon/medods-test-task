package repo

import (
	"context"

	"github.com/BabyJhon/medods-test-task/internal/entity"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Auth interface {
	CreateSession(ctx context.Context, session entity.Session) (uuid.UUID, error)
	GetSessionByID(ctx context.Context, sessionID uuid.UUID) (entity.Session, error)
	RevokeToken(ctx context.Context, sessionID uuid.UUID) error
	RefreshTokens(ctx context.Context, oldSession, newSession entity.Session) (uuid.UUID, error)
	GetAllSessions(ctx context.Context) ([]*entity.Session, error) 
}

type Repository struct {
	Auth
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		Auth: NewAuthRepo(db),
	}
}
