// internal/models/numero.go
package models

import "time"

// Numero representa um único número dentro de uma rifa.
type Numero struct {
	ID                int       `json:"id"`
	RifaID            int       `json:"rifa_id"`
	Numero            int       `json:"numero"`
	Status            string    `json:"status"`
	NomeComprador     *string   `json:"nome_comprador,omitempty"` // Ponteiro para aceitar NULL do banco
	EmailComprador    *string   `json:"email_comprador,omitempty"`
	TelefoneComprador *string   `json:"telefone_comprador,omitempty"`
	CreatedAt         time.Time `json:"-"` // Omitir no JSON de resposta por ser detalhe interno
	UpdatedAt         time.Time `json:"-"` // Omitir no JSON de resposta
}
