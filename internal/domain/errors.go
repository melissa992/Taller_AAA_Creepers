package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("usuario no encontrado")
	ErrUserAlreadyExists    = errors.New("el nombre de usuario o correo ya está registrado")
	ErrInvalidCredentials  = errors.New("credenciales inválidas")
	ErrSessionNotFound     = errors.New("sesión no encontrada")
	ErrSessionExpired      = errors.New("la sesión ha expirado")
	ErrVehicleNotFound     = errors.New("vehículo no encontrado")
	ErrUnauthorized        = errors.New("no autorizado")
	ErrForbidden           = errors.New("acceso denegado para su rol")
	ErrInvalidInput        = errors.New("datos de entrada inválidos")
)
