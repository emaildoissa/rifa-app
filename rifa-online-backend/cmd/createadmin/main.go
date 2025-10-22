// cmd/createadmin/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os" // Importe o 'os' para ler argumentos do terminal
	"rifa-online-backend/internal/database"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// 1. Carregar o .env para pegar a DATABASE_URL
	if err := godotenv.Load(".env"); err != nil { // Sobe dois níveis para achar o .env
		log.Fatal("Arquivo .env não encontrado. Rode o script da raiz do projeto.")
	}

	// 2. Conectar ao banco
	database.InitDB()
	defer database.CloseDB()

	// 3. Pegar email e senha dos argumentos do terminal
	if len(os.Args) != 3 {
		fmt.Println("Uso: go run ./cmd/createadmin [email] [senha]")
		os.Exit(1)
	}
	email := os.Args[1]
	password := os.Args[2]

	// 4. Criptografar (hash) a senha
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Erro ao gerar hash da senha: %v", err)
	}

	// 5. Inserir no banco de dados
	_, err = database.DB.Exec(context.Background(),
		"INSERT INTO users (email, password_hash) VALUES ($1, $2)",
		email, string(hash),
	)

	if err != nil {
		log.Fatalf("Erro ao inserir admin no banco: %v", err)
	}

	fmt.Printf("Usuário admin '%s' criado com sucesso!\n", email)
}
