package domain

import "time"

// Role define los roles del sistema según la especificación del taller.
type Role string

const (
	RoleVisitor Role = "Visitor"
	RoleSeller  Role = "Seller"
	RoleAdmin   Role = "Admin"
)

// User representa al usuario en el sistema con los campos requeridos: ID, Username, PasswordHash, Role.
type User struct {
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
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
