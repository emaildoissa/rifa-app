// internal/models/webhook.go
package models

// AsaasWebhookPayload é a estrutura principal do evento recebido do Asaas.
type AsaasWebhookPayload struct {
	Event   string      `json:"event"`
	Payment PaymentInfo `json:"payment"`
}

// PaymentInfo contém os detalhes do pagamento dentro do payload do webhook.
type PaymentInfo struct {
	ID     string  `json:"id"`
	Status string  `json:"status"`
	Value  float64 `json:"value"`
}
