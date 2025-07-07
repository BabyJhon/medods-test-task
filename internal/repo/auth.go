package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/BabyJhon/medods-test-task/internal/entity"
	"github.com/BabyJhon/medods-test-task/pkg/postgres"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuthRepo struct {
	db *pgxpool.Pool
}

func NewAuthRepo(db *pgxpool.Pool) *AuthRepo {
	return &AuthRepo{
		db: db,
	}
}

func (r *AuthRepo) CreateSession(ctx context.Context, session entity.Session) (uuid.UUID, error) {
	var id uuid.UUID
	query := fmt.Sprintf("INSERT INTO %s (id,  user_id,  refresh_hash, user_agent, ip, created_at,  expires_at, is_revoked) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", postgres.SessionTable)

	row := r.db.QueryRow(ctx, query, session.ID, session.UserId, session.RefreshHash, session.UserAgent, session.IP, session.CreatedAt, session.ExpiresAt, session.IsRevorked)
	if err := row.Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func (r *AuthRepo) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (entity.Session, error) {
	var session entity.Session

	query := fmt.Sprintf("SELECT * FROM %s WHERE id = $1", postgres.SessionTable)

	row := r.db.QueryRow(ctx, query, sessionID)
	if err := row.Scan(&session.ID, &session.UserId, &session.RefreshHash, &session.UserAgent, &session.IP, &session.CreatedAt, &session.ExpiresAt, &session.IsRevorked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Session{}, errors.New("session not found")
		}
		return entity.Session{}, err
	}

	return session, nil
}

func (r *AuthRepo) RevokeToken(ctx context.Context, sessionID uuid.UUID) error {
	revokeToken := fmt.Sprintf("UPDATE %s SET is_revoked = true WHERE id = $1", postgres.SessionTable)
	result, err := r.db.Exec(ctx, revokeToken, sessionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return errors.New("session not found or already revoked")
	}

	return nil
}


func (r *AuthRepo) GetAllSessions(ctx context.Context) ([]*entity.Session, error) {
	var sessions []*entity.Session
	query := fmt.Sprintf("SELECT * FROM %s WHERE is_revoked = false", postgres.SessionTable)

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var session entity.Session
		if err := rows.Scan(&session.ID, &session.UserId, &session.RefreshHash, &session.UserAgent, &session.IP, &session.CreatedAt, &session.ExpiresAt, &session.IsRevorked); err != nil {
			return nil, err
		}
		sessions = append(sessions, &session)
	}

	return sessions, nil
}

func (r *AuthRepo) RefreshTokens(ctx context.Context, oldSession, newSession entity.Session) (uuid.UUID, error) {
	var id uuid.UUID
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return id, err
	}

	revokeToken := fmt.Sprintf("UPDATE %s SET is_revoked = true WHERE id = $1", postgres.SessionTable)
	result, err := tx.Exec(ctx, revokeToken, oldSession.ID)
	if err != nil {
		tx.Rollback(ctx)
		return id, err
	}

	if result.RowsAffected() == 0 {
		tx.Rollback(ctx)
		return id, errors.New("session not found or already revoked")
	}

	query := fmt.Sprintf("INSERT INTO %s (id,  user_id,  refresh_hash, user_agent, ip, created_at,  expires_at, is_revoked) values ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id", postgres.SessionTable)

	row := tx.QueryRow(ctx, query, newSession.ID, newSession.UserId, newSession.RefreshHash, newSession.UserAgent, newSession.IP, newSession.CreatedAt, newSession.ExpiresAt, newSession.IsRevorked)
	if err := row.Scan(&id); err != nil {
		tx.Rollback(ctx)
		return id, err
	}

	return id, tx.Commit(ctx)
}
