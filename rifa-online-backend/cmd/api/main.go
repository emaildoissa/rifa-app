// cmd/api/main.go
package main

import (
	"log"
	"rifa-online-backend/internal/database"
	"rifa-online-backend/internal/handlers"
	"rifa-online-backend/internal/middleware" // <-- 1. IMPORTE O MIDDLEWARE

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Aviso: arquivo .env não encontrado.")
	}

	database.InitDB()
	defer database.CloseDB()

	router := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:5173",
		"http://191.252.223.221:8081",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"} // Permite o cabeçalho de Authorization
	router.Use(cors.New(config))

	// --- 2. DIVIDIR A API EM GRUPOS ---

	// Grupo /v1 (rotas públicas)
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "UP"})
		})

		// Rota de login pública
		v1.POST("/login", handlers.Login)

		// Rotas públicas de rifas
		v1.GET("/rifas", handlers.GetAllRifas)
		v1.GET("/rifas/:id", handlers.GetRifaByID)
		v1.POST("/rifas/:id/reservar", handlers.ReservarNumeros)

		// Rota de webhook
		//v1.POST("/webhooks/asaas", handlers.AsaasWebhookHandler)
	}

	// Grupo /admin (rotas protegidas)
	// --- 3. CRIAR O GRUPO DE ADMIN E APLICAR O MIDDLEWARE ---
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware()) // <-- O SEGURANÇA FICA NA PORTA DESTE GRUPO
	{
		// GET /api/v1/admin/rifas
		admin.GET("/rifas", handlers.GetAdminAllRifas)

		// POST /api/v1/admin/rifas
		admin.POST("/rifas", handlers.CreateRifa) // <-- MOVEMOS PARA CÁ

		// PUT /api/v1/admin/rifas/:id
		admin.PUT("/rifas/:id", handlers.UpdateRifa) // <-- MOVEMOS PARA CÁ

		// DELETE /api/v1/admin/rifas/:id
		admin.DELETE("/rifas/:id", handlers.DeleteRifa) // <-- MOVEMOS PARA CÁ
	}

	log.Println("Servidor iniciado na porta 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
