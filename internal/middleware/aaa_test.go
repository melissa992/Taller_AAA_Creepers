package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"automarket/internal/domain"
	"automarket/internal/middleware"
	"automarket/internal/repository/memory"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAAAWorkflow(t *testing.T) {
	sessionRepo := memory.NewSessionRepository()
	auditRepo := memory.NewAuditRepository(100)

	// Crear sesión válida en memoria para un usuario VENDEDOR
	testToken := "valid-test-token-1234567890abcdef"
	session := &domain.Session{
		Token:     testToken,
		UserID:    "usr-seller-01",
		Username:  "carlos_seller",
		Role:      domain.RoleSeller,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	_ = sessionRepo.Create(context.Background(), session)

	// Configurar router con los 3 middlewares AAA
	r := gin.New()
	r.Use(middleware.Accounting(auditRepo))
	r.Use(middleware.Authenticate(sessionRepo))

	// Ruta que requiere autenticación y rol ADMIN
	r.GET("/admin-only", middleware.RequireRoles(domain.RoleAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "Admin OK")
	})

	// Ruta que permite SELLER y ADMIN
	r.GET("/seller-resource", middleware.RequireRoles(domain.RoleSeller, domain.RoleAdmin), func(c *gin.Context) {
		c.String(http.StatusOK, "Seller OK")
	})

	// 1. Probar acceso anónimo a ruta protegida -> Debe denegar (401 o redirect)
	{
		req, _ := http.NewRequest(http.MethodGet, "/seller-resource", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Esperaba 401 Unauthorized para usuario anónimo, obtenido: %d", w.Code)
		}
	}

	// 2. Probar acceso autenticado con rol insuficiente (SELLER intenta entrar a /admin-only) -> 403 Forbidden
	{
		req, _ := http.NewRequest(http.MethodGet, "/admin-only", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.SessionCookieName,
			Value: testToken,
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Esperaba 403 Forbidden para rol SELLER en ruta ADMIN, obtenido: %d", w.Code)
		}
	}

	// 3. Probar acceso autenticado con rol correcto (SELLER a /seller-resource) -> 200 OK
	{
		req, _ := http.NewRequest(http.MethodGet, "/seller-resource", nil)
		req.AddCookie(&http.Cookie{
			Name:  middleware.SessionCookieName,
			Value: testToken,
		})
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Esperaba 200 OK para rol autorizado, obtenido: %d", w.Code)
		}
	}

	// 4. Verificar ACCOUNTING: Los 3 eventos deben estar registrados con Método, URL y UserID
	logs, err := auditRepo.ListRecent(context.Background(), 10)
	if err != nil {
		t.Fatalf("Error consultando logs de auditoría: %v", err)
	}

	if len(logs) != 3 {
		t.Fatalf("Se esperaban 3 logs registrados por Accounting, encontrados: %d", len(logs))
	}

	// El log más reciente debe ser el del acceso exitoso de carlos_seller
	lastLog := logs[0]
	if lastLog.UserID != "usr-seller-01" {
		t.Errorf("Esperaba UserID 'usr-seller-01' en último log, obtenido: %s", lastLog.UserID)
	}
	if lastLog.Role != "SELLER" {
		t.Errorf("Esperaba Role 'SELLER' en último log, obtenido: %s", lastLog.Role)
	}
	if lastLog.Status != 200 {
		t.Errorf("Esperaba Status 200 en último log, obtenido: %d", lastLog.Status)
	}
}
