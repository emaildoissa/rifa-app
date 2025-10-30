// internal/models/admin.go
package models

import "time"

// PagamentoPendente é a estrutura de dados para o dashboard do admin.
type PagamentoPendente struct {
	PagamentoID       int       `json:"pagamento_id"`
	RifaID            int       `json:"rifa_id"`
	RifaTitulo        string    `json:"rifa_titulo"`
	ValorTotal        float64   `json:"valor_total"`
	NomeComprador     string    `json:"nome_comprador"`
	EmailComprador    string    `json:"email_comprador"`
	TelefoneComprador string    `json:"telefone_comprador"`
	DataReserva       time.Time `json:"data_reserva"`
	Numeros           []int     `json:"numeros"` // Array de números reservados
}

type ParticipanteInfo struct {
	NomeComprador     string `json:"nome_comprador"`
	EmailComprador    string `json:"email_comprador"`
	TelefoneComprador string `json:"telefone_comprador"`
	Numeros           []int  `json:"numeros"` // Array de números pagos por este comprador
}

type WinnerInfo struct {
	NumeroSorteado    int     `json:"numero_sorteado"`
	NomeComprador     *string `json:"nome_comprador"`
	EmailComprador    *string `json:"email_comprador"`
	TelefoneComprador *string `json:"telefone_comprador"`
}
