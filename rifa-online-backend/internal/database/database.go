// internal/database/database.go
package database

import (
	"context"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

// InitDB inicializa a pool de conexões com o banco de dados PostgreSQL.
func InitDB() {
	// Pega a URL do banco de dados do arquivo .env
	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		log.Fatal("DATABASE_URL não foi definida no arquivo .env")
	}

	// Tenta conectar ao banco de dados
	var err error
	DB, err = pgxpool.New(context.Background(), databaseUrl)
	if err != nil {
		log.Fatalf("Não foi possível conectar ao banco de dados: %v\n", err)
	}

	// Testa a conexão para garantir que tudo está OK
	err = DB.Ping(context.Background())
	if err != nil {
		log.Fatalf("Não foi possível pingar o banco de dados: %v\n", err)
	}

	log.Println("Conectado ao banco de dados com sucesso!")
}

// CloseDB fecha a conexão com o banco de dados.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Conexão com o banco de dados fechada.")
	}
}
