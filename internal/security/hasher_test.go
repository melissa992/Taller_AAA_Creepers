package security_test

import (
	"testing"

	"automarket/internal/security"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "SecretPass123!"

	hash := security.HashPassword(password)
	if hash == "" {
		t.Fatal("El hash generado no debe estar vacío")
	}

	if !security.CheckPassword(password, hash) {
		t.Error("La verificación de la contraseña correcta falló")
	}

	if security.CheckPassword("WrongPassword", hash) {
		t.Error("La verificación aceptó una contraseña incorrecta")
	}
}

func TestSessionTokenGeneration(t *testing.T) {
	token1, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("Error generando token: %v", err)
	}

	token2, err := security.GenerateSessionToken()
	if err != nil {
		t.Fatalf("Error generando token: %v", err)
	}

	if len(token1) != 64 {
		t.Errorf("Se esperaba token de 64 caracteres hex, obtenido %d", len(token1))
	}

	if token1 == token2 {
		t.Error("Tokens consecutivos generados no deben ser idénticos")
	}
}
