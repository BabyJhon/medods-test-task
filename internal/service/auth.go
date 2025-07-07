package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/BabyJhon/medods-test-task/internal/entity"
	"github.com/BabyJhon/medods-test-task/internal/repo"
	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

const (
	bcryptCost      = 10
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 48 * time.Hour
)

type AuthService struct {
	repo repo.Auth
}

func NewAuthService(repo repo.Auth) *AuthService {
	return &AuthService{
		repo: repo,
	}
}

func (s *AuthService) generateAccessToken(userID, sessionID uuid.UUID, userAgent string, clientIP net.IP) (string, error) {
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS512, entity.Claimes{
		SessionID: sessionID,
		UserAgent: userAgent,
		IP:        clientIP,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})
	return accessToken.SignedString([]byte(os.Getenv("SIGNING_KEY")))
}

func (s *AuthService) generateRefreshToken() ([]byte, error) {
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("token generation error: %w", err)
	}

	return tokenBytes, nil
}

func (s *AuthService) CreateTokens(ctx context.Context, guid uuid.UUID, userAgent string, clientIP net.IP) (string, string, error) {
	refreshTokenBytes, err := s.generateRefreshToken()
	if err != nil {
		return "", "", err
	}

	BcryptBytes, err := bcrypt.GenerateFromPassword(refreshTokenBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	sessionID, err := uuid.DefaultGenerator.NewV4()
	if err != nil {
		return "", "", err
	}
	accessToken, err := s.generateAccessToken(guid, sessionID, userAgent, clientIP)
	if err != nil {
		return "", "", err
	}

	session := entity.Session{
		ID:          sessionID,
		UserId:      guid,
		RefreshHash: string(BcryptBytes),
		UserAgent:   userAgent,
		IP:          clientIP,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(refreshTokenTTL),
		IsRevorked:  false,
	}
	_, err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return "", "", err
	}

	refreshToken := base64.RawURLEncoding.EncodeToString(refreshTokenBytes)
	return accessToken, refreshToken, nil
}

func (a *AuthService) Parsetoken(accessToken string) (*entity.Claimes, error) {
	token, err := jwt.ParseWithClaims(accessToken, &entity.Claimes{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(os.Getenv("SIGNING_KEY")), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*entity.Claimes)
	if !ok {
		return nil, errors.New("token claims are not of type *Claimes")
	}

	return claims, nil
}

func (a *AuthService) GetSession(ctx context.Context, token entity.Claimes) (entity.Session, error) {
	session, err := a.repo.GetSessionByID(ctx, token.SessionID)
	if err != nil {
		return entity.Session{}, err
	}
	if session.IsRevorked {
		return entity.Session{}, errors.New("token is revoked")
	}

	return session, nil
}

func (a *AuthService) RevokeToken(ctx context.Context, sessionID uuid.UUID) error {
	err := a.repo.RevokeToken(ctx, sessionID)
	return err
}

func (a *AuthService) RefreshTokens(ctx context.Context, accessToken entity.Claimes, base64RefreshToken string, userAgent string, IP net.IP) (string, string, error) {

	decodedRefreshToken, err := base64.RawURLEncoding.DecodeString(base64RefreshToken)
	if err != nil {
		return "", "", err
	}
	sessions, err := a.repo.GetAllSessions(ctx)
	if err != nil {
		return "", "", err
	}
	var session entity.Session
	for i := 0; i < len(sessions); i++ {
		if bcrypt.CompareHashAndPassword([]byte(sessions[i].RefreshHash), decodedRefreshToken) == nil {
			session = *sessions[i]
		}
	}

	if session.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh token is expired")
	}

	if session.ID != accessToken.SessionID {
		return "", "", errors.New("the tokens were not created together")
	}

	if userAgent != session.UserAgent || userAgent != accessToken.UserAgent {
		err := a.repo.RevokeToken(ctx, session.ID)
		if err != nil {
			return "", "", err
		}
		return "", "", errors.New("user-agent has been changed")
	}

	if bytes.Equal(IP, session.IP) {
		message := fmt.Sprintf("received ip %s, expected %s", IP, session.IP)
		err := SendWebhook(WebhookPayload{Event: "wrong ip",
			Message: message,
		})
		if err != nil {
			return "", "", err
		}
	}
	if bytes.Equal(IP, accessToken.IP) {
		message := fmt.Sprintf("received ip %s, expected %s", IP, accessToken.IP)
		err := SendWebhook(WebhookPayload{Event: "wrong ip",
			Message: message,
		})
		if err != nil {
			return "", "", err
		}
	}

	newSessionID, err := uuid.DefaultGenerator.NewV4()
	if err != nil {
		return "", "", err
	}
	newAccessToken, err := a.generateAccessToken(session.UserId, newSessionID, userAgent, IP)
	if err != nil {
		return "", "", err
	}
	newRefreshTokenBytes, err := a.generateRefreshToken()
	if err != nil {
		return "", "", err
	}
	newBcryptBytes, err := bcrypt.GenerateFromPassword(newRefreshTokenBytes, bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	newSession := entity.Session{
		ID:          newSessionID,
		UserId:      session.UserId,
		RefreshHash: string(newBcryptBytes),
		UserAgent:   userAgent,
		IP:          IP,
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(refreshTokenTTL),
		IsRevorked:  false,
	}
	_, err = a.repo.RefreshTokens(ctx, session, newSession)
	if err != nil {
		return "", "", err
	}
	newRefreshToken := base64.RawURLEncoding.EncodeToString(newRefreshTokenBytes)
	return newAccessToken, newRefreshToken, nil
}
