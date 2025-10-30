// cmd/api/main.go
package main

import (
	"log"
	"rifa-online-backend/internal/database"
	"rifa-online-backend/internal/handlers"
	"rifa-online-backend/internal/middleware"

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
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Grupo /v1 (rotas públicas)
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "UP"})
		})

		v1.POST("/login", handlers.Login)

		v1.GET("/rifas", handlers.GetAllRifas)
		v1.GET("/rifas/:id", handlers.GetRifaByID)
		v1.POST("/rifas/:id/reservar", handlers.ReservarNumeros)

		// --- ROTA DO WEBHOOK REMOVIDA ---
		// v1.POST("/webhooks/asaas", handlers.AsaasWebhookHandler)
	}

	// Grupo /admin (rotas protegidas)
	admin := router.Group("/api/v1/admin")
	admin.Use(middleware.AuthMiddleware()) // <-- O SEGURANÇA FICA NA PORTA
	{
		// Rotas de gestão de Rifas
		admin.GET("/rifas", handlers.GetAdminAllRifas)
		admin.POST("/rifas", handlers.CreateRifa)
		admin.PUT("/rifas/:id", handlers.UpdateRifa)
		admin.DELETE("/rifas/:id", handlers.DeleteRifa)
		admin.GET("/rifas/:id/participantes", handlers.GetParticipantesPorRifa) // <-- ADICIONE ESTA LINHA

		// --- NOVAS ROTAS DE GESTÃO DE PAGAMENTOS ---
		admin.GET("/pagamentos/pendentes", handlers.GetPendingPagamentos)
		admin.POST("/pagamentos/:id/aprovar", handlers.AprovarPagamento)
		admin.POST("/pagamentos/:id/liberar", handlers.LiberarPagamento)

		admin.POST("/rifas/:id/sortear", handlers.SortearRifa)

	}

	log.Println("Servidor iniciado na porta 8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Falha ao iniciar o servidor: %v", err)
	}
}
