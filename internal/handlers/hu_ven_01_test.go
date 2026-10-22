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

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHU_VEN_01_Workflow(t *testing.T) {
	repo := repository.GetInstance()

	r := gin.New()
	r.LoadHTMLGlob("../../templates/*")

	r.Use(middleware.Accounting(repo))
	r.Use(middleware.Authenticate(repo))

	authHandler := handlers.NewAuthHandler(repo, repo)
	vehicleHandler := handlers.NewVehicleHandler(repo)

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	sellerGroup := r.Group("/")
	sellerGroup.Use(middleware.RequireAuth())
	sellerGroup.Use(middleware.RequireRoles(domain.RoleSeller))
	{
		sellerGroup.GET("/dashboard", vehicleHandler.Dashboard)
		sellerGroup.POST("/vehicles", vehicleHandler.Create)
	}

	// 1. Registro de nuevo usuario (debe asignarse rol 'Seller' y contraseña con crypto/sha256)
	testUsername := "vendedor_test_01"
	testPassword := "Vendedor123!"

	formRegister := url.Values{}
	formRegister.Set("username", testUsername)
	formRegister.Set("password", testPassword)
	formRegister.Set("confirm_password", testPassword)

	reqReg, _ := http.NewRequest(http.MethodPost, "/register", strings.NewReader(formRegister.Encode()))
	reqReg.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wReg := httptest.NewRecorder()
	r.ServeHTTP(wReg, reqReg)

	if wReg.Code != http.StatusSeeOther {
		t.Fatalf("Se esperaba 303 SeeOther tras registro exitoso, obtenido: %d", wReg.Code)
	}

	// Verificar que el usuario en el repositorio tenga rol 'Seller'
	userCreated, err := repo.GetUserByUsername(context.Background(), testUsername)
	if err != nil {
		t.Fatalf("Usuario no encontrado en repositorio: %v", err)
	}
	if userCreated.Role != domain.RoleSeller {
		t.Errorf("Se esperaba rol Seller, obtenido: %s", userCreated.Role)
	}

	// 2. Login con credenciales válidas -> Debe devolver cookie session_token
	formLogin := url.Values{}
	formLogin.Set("username", testUsername)
	formLogin.Set("password", testPassword)

	reqLogin, _ := http.NewRequest(http.MethodPost, "/login", strings.NewReader(formLogin.Encode()))
	reqLogin.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusSeeOther {
		t.Fatalf("Se esperaba 303 SeeOther tras login, obtenido: %d", wLogin.Code)
	}

	cookies := wLogin.Result().Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == middleware.SessionCookieName {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("No se encontró cookie session_token tras login")
	}

	// 3. Intento de crear vehículo SIN autenticación -> Debe redirigir a /login (RequireAuth)
	formVeh := url.Values{}
	formVeh.Set("title", "Mazda CX-5 2023")
	formVeh.Set("price_formatted", "$28,000 USD")
	formVeh.Set("year_mileage", "2023 • 12,000 km")
	formVeh.Set("location", "Medellín, Colombia")
	formVeh.Set("whatsapp_number", "573111223344")

	reqUnauth, _ := http.NewRequest(http.MethodPost, "/vehicles", strings.NewReader(formVeh.Encode()))
	reqUnauth.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	wUnauth := httptest.NewRecorder()
	r.ServeHTTP(wUnauth, reqUnauth)

	if wUnauth.Code != http.StatusSeeOther || !strings.Contains(wUnauth.Header().Get("Location"), "/login") {
		t.Errorf("Petición sin auth debía redirigir a login, obtenido: %d", wUnauth.Code)
	}

	// 4. Creación de vehículo CON autenticación (Rol Seller) -> Debe quedar en estado 'Pending'
	reqVeh, _ := http.NewRequest(http.MethodPost, "/vehicles", strings.NewReader(formVeh.Encode()))
	reqVeh.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	reqVeh.AddCookie(sessionCookie)
	wVeh := httptest.NewRecorder()
	r.ServeHTTP(wVeh, reqVeh)

	if wVeh.Code != http.StatusSeeOther {
		t.Fatalf("Se esperaba 303 SeeOther al crear vehículo, obtenido: %d", wVeh.Code)
	}

	// Verificar en el repositorio que el vehículo fue creado con estado 'Pending' y el SellerID correspondiente
	sellerVehicles, err := repo.GetBySellerID(context.Background(), userCreated.ID)
	if err != nil || len(sellerVehicles) == 0 {
		t.Fatalf("No se encontraron vehículos para el vendedor: %v", err)
	}

	createdVeh := sellerVehicles[0]
	if createdVeh.Status != domain.StatusPending {
		t.Errorf("Se esperaba estado 'Pending' para el nuevo vehículo, obtenido: %s", createdVeh.Status)
	}
	if createdVeh.SellerID != userCreated.ID {
		t.Errorf("Se esperaba SellerID %s, obtenido: %s", userCreated.ID, createdVeh.SellerID)
	}
	if createdVeh.Title != "Mazda CX-5 2023" {
		t.Errorf("Título inesperado: %s", createdVeh.Title)
	}
}
