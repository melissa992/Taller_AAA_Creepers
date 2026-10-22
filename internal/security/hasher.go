package security

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// HashPassword calcula el hash SHA-256 de una contraseña.
// Cumple con la restricción de usar 'crypto/sha256' de la biblioteca estándar de Go.
func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// CheckPassword compara de forma segura en tiempo constante el hash calculado con el almacenado,
// protegiendo contra ataques de análisis de tiempo (Timing Attacks).
func CheckPassword(password, expectedHash string) bool {
	computedHash := HashPassword(password)
	return subtle.ConstantTimeCompare([]byte(computedHash), []byte(expectedHash)) == 1
}

// GenerateSessionToken genera un token criptográfico seguro de 32 bytes (64 caracteres hexadecimales)
// para mapear la sesión del usuario en memoria.
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token de sesión: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateID genera un identificador único seguro con prefijo.
func GenerateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", prefix, hex.EncodeToString(b))
}
