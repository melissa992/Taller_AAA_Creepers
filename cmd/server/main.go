package main

import (
	"log"
	"os"

	"automarket/internal/domain"
	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("[INIT] Inicializando AutoMarket")

	// 1. Repositorio en memoria (Singleton, thread-safe con sync.RWMutex)
	repo := repository.GetInstance()

	// 2. Configurar motor Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 3. Estáticos y Plantillas
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	// 4. Middlewares Globales
	// Accounting: Intercepta y registra todas las peticiones con trazabilidad e integridad SHA-256
	router.Use(middleware.Accounting(repo))
	// Authentication: Valida cookie "session_token" e inyecta user_id y rol en gin.Context
	router.Use(middleware.Authenticate(repo))

	// 5. Inyección de Handlers
	catalogHandler := handlers.NewCatalogHandler(repo)
	authHandler := handlers.NewAuthHandler(repo, repo)
	vehicleHandler := handlers.NewVehicleHandler(repo)
	adminHandler := handlers.NewAdminHandler(repo, repo, repo)

	// 6. Rutas Públicas (HU-VIS-01 y Autenticación)
	router.GET("/", catalogHandler.Index)

	router.GET("/register", authHandler.ShowRegister)
	router.POST("/register", authHandler.Register)

	router.GET("/login", authHandler.ShowLogin)
	router.POST("/login", authHandler.Login)
	router.POST("/logout", authHandler.Logout)

	// 7. Rutas Protegidas por Autenticación y Autorización (HU-VEN-01 - Solo 'Seller')
	sellerGroup := router.Group("/")
	sellerGroup.Use(middleware.RequireAuth())
	sellerGroup.Use(middleware.RequireRoles(domain.RoleSeller))
	{
		sellerGroup.GET("/dashboard", vehicleHandler.Dashboard)
		sellerGroup.POST("/vehicles", vehicleHandler.Create)
	}

	// 8. Rutas Protegidas por Autenticación y Autorización (HU-ADM-01 - Solo 'Admin')
	adminGroup := router.Group("/admin")
	adminGroup.Use(middleware.RequireAuth())
	adminGroup.Use(middleware.RequireRoles(domain.RoleAdmin))
	{
		adminGroup.GET("/publicaciones", adminHandler.ShowPublications)
		adminGroup.POST("/publicaciones/:id/aprobar", adminHandler.ApprovePublication)
		adminGroup.POST("/publicaciones/:id/eliminar", adminHandler.DeletePublication)

		adminGroup.GET("/usuarios", adminHandler.ShowUsers)
		adminGroup.POST("/usuarios/:id/eliminar", adminHandler.DeleteUser)
	}

	// 9. Iniciar el servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER] AutoMarket ejecutándose en http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[FATAL] Error ejecutando el servidor: %v", err)
	}
}
