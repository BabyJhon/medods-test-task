package entity

import (
	"net"

	"github.com/gofrs/uuid"
	"github.com/golang-jwt/jwt/v5"
)

type Claimes struct {
	SessionID uuid.UUID `json:"session_id"`
	UserAgent string    `json:"user_agent"`
	IP        net.IP    `json:"ip"`
	jwt.RegisteredClaims
}
