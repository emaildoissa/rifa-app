// internal/models/rifa.go
package models

import "time"

// Rifa representa a estrutura de dados de uma rifa no sistema.
type Rifa struct {
	ID             int       `json:"id"`
	Titulo         string    `json:"titulo" binding:"required"`
	Descricao      string    `json:"descricao"`
	Premio         string    `json:"premio" binding:"required"`
	PrecoPorNumero float64   `json:"preco_por_numero" binding:"required,gt=0"`
	TotalNumeros   int       `json:"total_numeros" binding:"required,gt=0"`
	DataSorteio    time.Time `json:"data_sorteio"`
	Status         string    `json:"status"`
	NumeroSorteado int       `json:"numero_sorteado,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
type RifaSummary struct {
	ID              int     `json:"id"`
	Titulo          string  `json:"titulo"`
	Premio          string  `json:"premio"`
	PrecoPorNumero  float64 `json:"preco_por_numero"`
	Status          string  `json:"status"`
	NumerosVendidos int     `json:"numeros_vendidos"`
	TotalNumeros    int     `json:"total_numeros"`
}

type RifaUpdateInput struct {
	Titulo      string    `json:"titulo" binding:"required"`
	Descricao   string    `json:"descricao"`
	Premio      string    `json:"premio" binding:"required"`
	DataSorteio time.Time `json:"data_sorteio"`
	Status      string    `json:"status" binding:"required"`
}
