// internal/payment/asaas.go
package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"rifa-online-backend/internal/models"
	"time"
)

// Constantes da API
const asaasBaseURL = "https://sandbox.asaas.com/api/v3"

// AsaasChargeRequest é a estrutura para criar uma nova cobrança no Asaas
type AsaasChargeRequest struct {
	Customer          string  `json:"customer"`
	BillingType       string  `json:"billingType"`
	Value             float64 `json:"value"`
	DueDate           string  `json:"dueDate"`
	Description       string  `json:"description"`
	ExternalReference string  `json:"externalReference"` // Usaremos para nosso ID de pagamento
}

// AsaasChargeResponse é a estrutura para a resposta de criação de cobrança
type AsaasChargeResponse struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Value     float64 `json:"value"`
	PixQrCode struct {
		Payload        string `json:"payload"` // O "copia e cola" do PIX
		ExpirationDate string `json:"expirationDate"`
	} `json:"pixQrCode"`
	InvoiceURL string `json:"invoiceUrl"`
}

// AsaasCustomerRequest é para criar um novo cliente
type AsaasCustomerRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	CpfCnpj string `json:"cpfCnpj"` // O Asaas exige um CPF/CNPJ
}

// AsaasCustomerResponse é a resposta da API de clientes
type AsaasCustomerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AsaasCustomerListResponse struct {
	Data []AsaasCustomerResponse `json:"data"`
}

// CreatePixCharge cria uma cobrança PIX no Asaas
func CreatePixCharge(customerID string, rifa models.Rifa, valorTotal float64, pagamentoID int) (*AsaasChargeResponse, error) {
	apiKey := os.Getenv("ASAAS_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ASAAS_API_KEY não encontrada")
	}

	payload := AsaasChargeRequest{
		Customer:          customerID, // No Asaas, precisamos de um ID de cliente primeiro.
		BillingType:       "PIX",
		Value:             valorTotal,
		DueDate:           time.Now().Format("2006-01-02"),
		Description:       "Pagamento da Rifa: " + rifa.Titulo,
		ExternalReference: string(rune(pagamentoID)), // Vincula a cobrança ao nosso ID de pagamento interno
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", asaasBaseURL+"/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", apiKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Erro na API Asaas: Status %d", resp.StatusCode)
		// Aqui poderíamos ler o corpo da resposta para ver o erro detalhado
		return nil, errors.New("falha ao criar cobrança no Asaas")
	}

	var chargeResponse AsaasChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResponse); err != nil {
		return nil, err
	}

	return &chargeResponse, nil
}

func FindOrCreateCustomer(nome, email, cpf string) (string, error) {
	apiKey := os.Getenv("ASAAS_API_KEY")
	// (Lógica de validação da apiKey omitida por brevidade)

	// 1. Tenta buscar o cliente pelo CPF
	searchURL := asaasBaseURL + "/customers?cpfCnpj=" + cpf
	req, _ := http.NewRequest("GET", searchURL, nil)
	req.Header.Set("access_token", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var searchResult AsaasCustomerListResponse
	json.NewDecoder(resp.Body).Decode(&searchResult)

	if len(searchResult.Data) > 0 {
		return searchResult.Data[0].ID, nil // Cliente encontrado
	}

	// 2. Se não encontrou, cria um novo
	customerPayload := AsaasCustomerRequest{Name: nome, Email: email, CpfCnpj: cpf}
	jsonData, _ := json.Marshal(customerPayload)

	createURL := asaasBaseURL + "/customers"
	req, _ = http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", apiKey)

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("falha ao criar cliente no Asaas")
	}

	var newCustomer AsaasCustomerResponse
	json.NewDecoder(resp.Body).Decode(&newCustomer)

	return newCustomer.ID, nil
}
