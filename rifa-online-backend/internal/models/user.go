// internal/models/user.go
package models

// User representa um usuário no banco de dados.
type User struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"` // O '-' omite o hash da senha de qualquer JSON
}

// LoginInput é a estrutura para a requisição de login.
type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
