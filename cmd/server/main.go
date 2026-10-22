package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"automarket/internal/bootstrap"
	"automarket/internal/domain"
	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/repository/memory"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("[INIT] Inicializando AutoMarket (Arquitectura AAA)...")

	// 1. Instanciación de Repositorios en Memoria protegidos con sync.RWMutex
	userRepo := memory.NewUserRepository()
	sessionRepo := memory.NewSessionRepository()
	vehicleRepo := memory.NewVehicleRepository()
	auditRepo := memory.NewAuditRepository(1000)

	// 2. Cargar datos semilla (Admin, Vendedor, Comprador y vehículos iniciales)
	if err := bootstrap.SeedInitialData(userRepo, vehicleRepo); err != nil {
		log.Fatalf("[FATAL] Error cargando datos de prueba: %v", err)
	}
	log.Println("[INIT] Datos semilla cargados exitosamente (Admin, Vendedor, Comprador y Catálogo)")

	// 3. Configurar motor Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 4. Servir archivos estáticos (CSS, fuentes)
	router.Static("/static", "./web/static")

	// 5. Cargar plantillas HTML de forma recursiva y segura
	tmpl, err := loadTemplates("web/templates")
	if err != nil {
		log.Fatalf("[FATAL] Error cargando plantillas HTML: %v", err)
	}
	router.SetHTMLTemplate(tmpl)

	// 6. MIDDLEWARES GLOBALES (Arquitectura AAA RF-01)
	// Accounting: Intercepta y registra todas las peticiones (Método, URL, UserID, Timestamp, Latencia, Status)
	router.Use(middleware.Accounting(auditRepo))
	// Authentication: Valida la presencia de cookie segura de sesión y asocia la identidad al contexto
	router.Use(middleware.Authenticate(sessionRepo))

	// 7. CONTROLADORES (Inyección de Dependencias)
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo)
	vehicleHandler := handlers.NewVehicleHandler(vehicleRepo)
	dashboardHandler := handlers.NewDashboardHandler(vehicleRepo, auditRepo, userRepo)

	// 8. DEFINICIÓN DE RUTAS Y CONTROL DE ACCESO (RBAC)

	// Rutas Públicas (Navegación libre y autenticación)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/vehicles")
	})
	router.GET("/vehicles", vehicleHandler.List)
	router.GET("/vehicles/:id", vehicleHandler.Show)

	router.GET("/login", authHandler.ShowLogin)
	router.POST("/login", authHandler.Login)
	router.GET("/register", authHandler.ShowRegister)
	router.POST("/register", authHandler.Register)
	router.POST("/logout", authHandler.Logout)

	// Rutas Protegidas - Requieren Autenticación
	authRoutes := router.Group("/")
	authRoutes.Use(middleware.RequireAuth())
	{
		authRoutes.GET("/dashboard", dashboardHandler.Index)

		// Rutas Protegidas por Rol - Vendedores y Administradores
		sellerRoutes := authRoutes.Group("/")
		sellerRoutes.Use(middleware.RequireRoles(domain.RoleSeller, domain.RoleAdmin))
		{
			sellerRoutes.GET("/vehicles/new", vehicleHandler.ShowCreate)
			sellerRoutes.POST("/vehicles", vehicleHandler.Create)
			sellerRoutes.GET("/vehicles/:id/edit", vehicleHandler.ShowEdit)
			sellerRoutes.POST("/vehicles/:id/edit", vehicleHandler.Update)
			sellerRoutes.POST("/vehicles/:id/delete", vehicleHandler.Delete)
		}

		// Rutas Exclusivas de Administrador - Auditoría AAA (Accounting)
		adminRoutes := authRoutes.Group("/dashboard")
		adminRoutes.Use(middleware.RequireRoles(domain.RoleAdmin))
		{
			adminRoutes.GET("/audit", dashboardHandler.AuditLogs)
		}
	}

	// Manejo semántico de 404 Not Found
	router.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "errors/404.html", gin.H{
			"Title":   "404 - No Encontrado",
			"Message": "La ruta solicitada no existe en el servidor AutoMarket",
		})
	})

	// 9. Inicio del Servidor Web
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := ":" + port
	log.Printf("[SERVER] AutoMarket ejecutándose en http://localhost:%s\n", port)
	log.Println("[INFO] Cuentas demo disponibles: admin/Admin123!, vendedor1/Vendedor123!, comprador1/Comprador123!")

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("[FATAL] Error iniciando servidor HTTP: %v", err)
	}
}

// loadTemplates recorre el directorio de plantillas HTML y las compila en un conjunto unificado.
func loadTemplates(rootDir string) (*template.Template, error) {
	rootTmpl := template.New("")

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".html") {
			relPath, err := filepath.Rel(rootDir, path)
			if err != nil {
				return err
			}
			// Normalizar separadores de ruta a formato web
			relPath = filepath.ToSlash(relPath)

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			// Asociar la plantilla con su nombre relativo y procesar definiciones
			_, err = rootTmpl.New(relPath).Parse(string(content))
			if err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return rootTmpl, nil
}
