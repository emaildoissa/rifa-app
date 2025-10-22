// internal/middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware é o nosso "segurança" que protege as rotas.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Pegar o cabeçalho "Authorization"
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token de autorização não fornecido"})
			return
		}

		// 2. O cabeçalho deve estar no formato "Bearer <token>"
		// Vamos extrair apenas o token.
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Formato do token inválido"})
			return
		}

		// 3. Pegar nosso segredo do .env
		jwtSecret := os.Getenv("JWT_SECRET_KEY")
		if jwtSecret == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Chave de assinatura do servidor não configurada"})
			return
		}

		// 4. Validar o token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Verifica se o método de assinatura é o que esperamos (HS256)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("método de assinatura inesperado: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		})

		if err != nil {
			// Erros de validação (token expirado, assinatura inválida, etc.)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("Token inválido: %v", err)})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Opcional: Adicionar o ID do usuário ao contexto da requisição
			// para que os handlers futuros saibam quem fez a chamada.
			userID := claims["sub"]
			c.Set("userID", userID)

			// 5. Token é válido! Deixa a requisição continuar.
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		}
	}
}
