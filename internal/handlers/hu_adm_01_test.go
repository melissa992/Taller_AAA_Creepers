package handlers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"automarket/internal/domain"
	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/repository"

	"github.com/gin-gonic/gin"
)

func setupAdminTestRouter(repo *repository.MemoryRepository) *gin.Engine {
	r := gin.New()
	r.LoadHTMLGlob("../../templates/*")

	r.Use(middleware.Accounting(repo))
	r.Use(middleware.Authenticate(repo))

	authHandler := handlers.NewAuthHandler(repo, repo)
	adminHandler := handlers.NewAdminHandler(repo, repo, repo)

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.RequireAuth())
	adminGroup.Use(middleware.RequireRoles(domain.RoleAdmin))
	{
		adminGroup.GET("/publicaciones", adminHandler.ShowPublications)
		adminGroup.POST("/publicaciones/:id/aprobar", adminHandler.ApprovePublication)
		adminGroup.POST("/publicaciones/:id/eliminar", adminHandler.DeletePublication)
		adminGroup.GET("/usuarios", adminHandler.ShowUsers)
		adminGroup.POST("/usuarios/:id/eliminar", adminHandler.DeleteUser)
	}

	return r
}

func loginAndGetCookie(t *testing.T, r *gin.Engine, username, password string) *http.Cookie {
	t.Helper()
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)

	req, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("Login falló para %s: status %d", username, w.Code)
	}

	for _, c := range w.Result().Cookies() {
		if c.Name == middleware.SessionCookieName {
			return c
		}
	}
	t.Fatalf("No se emitió cookie de sesión para %s", username)
	return nil
}

func TestHU_ADM_01_And_RF_01_Workflow(t *testing.T) {
	repo := repository.GetInstance()
	ctx := context.Background()
	r := setupAdminTestRouter(repo)

	// Crear un vehículo pendiente de prueba para moderar
	pendingVeh := &domain.Vehicle{
		ID:             "veh-test-pending",
		ImageURL:       "/assets/images/car1.jpg",
		Title:          "Subaru WRX STI 2021",
		PriceFormatted: "$45,000 USD",
		YearMileage:    "2021 • 22,000 km",
		Location:       "Bogotá, Colombia",
		WhatsAppNumber: "573001112233",
		Status:         domain.StatusPending,
		SellerID:       "usr-seller-1",
	}
	if err := repo.Create(ctx, pendingVeh); err != nil {
		t.Fatalf("Error creando vehículo pendiente de prueba: %v", err)
	}

	// 1. Visitante no autenticado intenta aprobar directamente la ruta -> Debe ser rechazado (303 a /login)
	reqVis, _ := http.NewRequest(http.MethodPost, "/admin/publicaciones/veh-test-pending/aprobar", nil)
	wVis := httptest.NewRecorder()
	r.ServeHTTP(wVis, reqVis)
	if wVis.Code != http.StatusSeeOther || !strings.Contains(wVis.Header().Get("Location"), "/login") {
		t.Fatalf("Visitante no debía poder aprobar publicación, código obtenido: %d", wVis.Code)
	}

	// 2. Vendedor autenticado intenta aprobar o eliminar directamente la ruta -> Debe recibir 403 Forbidden
	sellerCookie := loginAndGetCookie(t, r, "carlos_autos", "Seller123!")
	reqSel, _ := http.NewRequest(http.MethodPost, "/admin/publicaciones/veh-test-pending/aprobar", nil)
	reqSel.AddCookie(sellerCookie)
	wSel := httptest.NewRecorder()
	r.ServeHTTP(wSel, reqSel)
	if wSel.Code != http.StatusForbidden {
		t.Fatalf("Vendedor debía recibir 403 Forbidden en ruta admin, obtuvo: %d", wSel.Code)
	}

	// Verificar que el vehículo sigue en estado Pending
	vCheck, _ := repo.GetByID(ctx, "veh-test-pending")
	if vCheck.Status != domain.StatusPending {
		t.Fatalf("El vehículo no debía cambiar de estado tras intentos no autorizados, estado actual: %s", vCheck.Status)
	}

	// 3. Administrador autenticado aprueba la publicación pendiente -> Cambia a Published
	adminCookie := loginAndGetCookie(t, r, "admin", "Admin123!")
	reqApp, _ := http.NewRequest(http.MethodPost, "/admin/publicaciones/veh-test-pending/aprobar", nil)
	reqApp.AddCookie(adminCookie)
	wApp := httptest.NewRecorder()
	r.ServeHTTP(wApp, reqApp)
	if wApp.Code != http.StatusSeeOther {
		t.Fatalf("Admin debía aprobar con redirección 303, obtuvo: %d", wApp.Code)
	}

	vApproved, err := repo.GetByID(ctx, "veh-test-pending")
	if err != nil || vApproved.Status != domain.StatusPublished {
		t.Fatalf("Se esperaba estado Published tras aprobación del admin, obtuvo: %v", vApproved)
	}

	// 4. Administrador elimina la publicación -> Desaparece del repositorio
	reqDelPub, _ := http.NewRequest(http.MethodPost, "/admin/publicaciones/veh-test-pending/eliminar", nil)
	reqDelPub.AddCookie(adminCookie)
	wDelPub := httptest.NewRecorder()
	r.ServeHTTP(wDelPub, reqDelPub)
	if wDelPub.Code != http.StatusSeeOther {
		t.Fatalf("Admin debía eliminar publicación con 303, obtuvo: %d", wDelPub.Code)
	}
	if _, err := repo.GetByID(ctx, "veh-test-pending"); err == nil {
		t.Fatal("La publicación eliminada aún existe en el repositorio")
	}

	// 5. Administrador elimina un usuario vendedor -> Se elimina el usuario y sus publicaciones en cascada
	tempUser := &domain.User{
		ID:           "usr-temp-delete",
		Username:     "vendedor_eliminar",
		PasswordHash: "hash",
		Role:         domain.RoleSeller,
	}
	_ = repo.CreateUser(ctx, tempUser)
	_ = repo.Create(ctx, &domain.Vehicle{
		ID:       "veh-cascade-1",
		Title:    "Auto de usuario eliminado",
		Status:   domain.StatusPublished,
		SellerID: tempUser.ID,
	})

	reqDelUsr, _ := http.NewRequest(http.MethodPost, "/admin/usuarios/"+tempUser.ID+"/eliminar", nil)
	reqDelUsr.AddCookie(adminCookie)
	wDelUsr := httptest.NewRecorder()
	r.ServeHTTP(wDelUsr, reqDelUsr)
	if wDelUsr.Code != http.StatusSeeOther {
		t.Fatalf("Admin debía eliminar usuario con 303, obtuvo: %d", wDelUsr.Code)
	}

	if _, err := repo.GetUserByID(ctx, tempUser.ID); err == nil {
		t.Fatal("El usuario eliminado aún existe en el repositorio")
	}
	if _, err := repo.GetByID(ctx, "veh-cascade-1"); err == nil {
		t.Fatal("La publicación del usuario eliminado no se eliminó en cascada")
	}

	// 6. Verificar integridad de la cadena de auditoría SHA-256 (RF-01 Accounting)
	ok, badSeq, err := repo.VerifyIntegrity(ctx)
	if !ok || err != nil {
		t.Fatalf("Fallo en VerifyIntegrity de auditoría en secuencia %d: %v", badSeq, err)
	}
}
