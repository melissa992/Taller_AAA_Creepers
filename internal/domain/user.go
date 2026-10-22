package domain

import "time"

// Role representa los niveles de autorización en el sistema AutoMarket.
type Role string

const (
	RoleAdmin  Role = "ADMIN"
	RoleSeller Role = "SELLER"
	RoleBuyer  Role = "BUYER"
)

// User representa la entidad pura de un usuario en el dominio.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Salt         string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// Clone produce una copia profunda para evitar condiciones de carrera por mutación externa.
func (u *User) Clone() *User {
	if u == nil {
		return nil
	}
	copyUser := *u
	return &copyUser
}
