// internal/handlers/auth_handler.go
package handlers

import (
	"context"
	"log"
	"net/http"
	"os"
	"rifa-online-backend/internal/database"
	"rifa-online-backend/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

func Login(c *gin.Context) {
	var input models.LoginInput

	// 1. Validar o JSON de entrada
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "E-mail ou senha inválidos"})
		return
	}

	// 2. Buscar o usuário no banco
	var user models.User
	err := database.DB.QueryRow(context.Background(),
		"SELECT id, email, password_hash FROM users WHERE email = $1",
		input.Email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("Tentativa de login falhou (usuário não encontrado): %s", input.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "E-mail ou senha inválidos"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}

	// 3. Comparar a senha enviada com o hash salvo
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		// Senha não confere
		log.Printf("Tentativa de login falhou (senha inválida): %s", input.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "E-mail ou senha inválidos"})
		return
	}

	// 4. Gerar o Token JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,                              // "Subject" (Quem é o usuário)
		"exp": time.Now().Add(time.Hour * 8).Unix(), // Expira em 8 horas
	})

	// 5. Assinar o token com nosso segredo do .env
	jwtSecret := os.Getenv("JWT_SECRET_KEY")
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar token"})
		return
	}

	// 6. Enviar o token para o cliente
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}
