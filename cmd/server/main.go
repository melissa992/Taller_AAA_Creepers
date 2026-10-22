package main

import (
	"log"
	"os"

	"automarket/internal/handlers"
	"automarket/internal/middleware"
	"automarket/internal/repository"

	"github.com/gin-gonic/gin"
)

func main() {
	log.Println("[INIT] Inicializando AutoMarket - HU-VIS-01...")

	// 1. Obtener repositorio en memoria (Singleton, thread-safe con sync.RWMutex)
	repo := repository.GetInstance()

	// 2. Configurar motor Gin
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// 3. Configuración solicitada de archivos estáticos y plantillas:
	// - router.Static("/assets", "./assets")
	// - router.LoadHTMLGlob("templates/*")
	router.Static("/assets", "./assets")
	router.LoadHTMLGlob("templates/*")

	// 4. Middleware de Accounting (AAA)
	// Intercepta y registra: Método, URL, UserID y Timestamp
	router.Use(middleware.Accounting(repo))

	// 5. Inyección del handler del catálogo público
	catalogHandler := handlers.NewCatalogHandler(repo)

	// 6. Handler GET '/' que consulta el repositorio (solo "Published") y lo pasa a la plantilla
	router.GET("/", catalogHandler.Index)

	// 7. Iniciar el servidor
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("[SERVER] AutoMarket corriendo en http://localhost:%s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[FATAL] Error ejecutando el servidor: %v", err)
	}
}
