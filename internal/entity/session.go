package entity

import (
	"net"
	"time"

	"github.com/gofrs/uuid"
)

type Session struct {
	ID          uuid.UUID `db:"id"`
	UserId      uuid.UUID `json:"user_id" db:"user_id"`
	RefreshHash string    `db:"refresh_hash"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	IP          net.IP    `json:"ip" db:"ip"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiresAt   time.Time `db:"expires_at"`
	IsRevorked  bool      `db:"is_revorked"`
}
