package domain

import "time"

// Session representa una sesión de usuario activa en memoria mapeada a una cookie segura.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Role      Role      `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired verifica si la sesión ha sobrepasado su tiempo de vida.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// Clone produce una copia profunda de la sesión.
func (s *Session) Clone() *Session {
	if s == nil {
		return nil
	}
	copySession := *s
	return &copySession
}
