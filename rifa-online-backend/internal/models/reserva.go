// internal/models/reserva.go
package models

// ReservaInput é a estrutura para a requisição de reserva de números.
type ReservaInput struct {
	NomeComprador     string `json:"nome_comprador" binding:"required"`
	EmailComprador    string `json:"email_comprador" binding:"required,email"`
	TelefoneComprador string `json:"telefone_comprador" binding:"required"`
	CpfCnpj           string `json:"cpf_cnpj" binding:"required"`
	Numeros           []int  `json:"numeros" binding:"required,min=1"`
}
