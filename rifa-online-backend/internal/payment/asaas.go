// internal/payment/asaas.go
package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt" // Import fmt para Sprintf nos logs
	"io"  // Import io para ler o corpo da resposta
	"log"
	"net/http"
	"os"
	"rifa-online-backend/internal/models"
	"strconv" // Import strconv para Itoa
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
	ExternalReference string  `json:"externalReference"`
}

// AsaasChargeResponse é a estrutura para a resposta de criação de cobrança
type AsaasChargeResponse struct {
	ID        string  `json:"id"`
	Status    string  `json:"status"`
	Value     float64 `json:"value"`
	PixQrCode struct {
		Payload        string `json:"payload"`
		ExpirationDate string `json:"expirationDate"`
	} `json:"pixQrCode"`
	InvoiceURL string `json:"invoiceUrl"`
}

// CreatePixCharge cria uma cobrança PIX no Asaas
func CreatePixCharge(customerID string, rifa models.Rifa, valorTotal float64, pagamentoID int) (*AsaasChargeResponse, error) {
	apiKey := os.Getenv("ASAAS_API_KEY")
	if apiKey == "" {
		return nil, errors.New("ASAAS_API_KEY não encontrada")
	}

	payload := AsaasChargeRequest{
		Customer:          customerID,
		BillingType:       "PIX",
		Value:             valorTotal,
		DueDate:           time.Now().Format("2006-01-02"),
		Description:       "Pagamento da Rifa: " + rifa.Titulo,
		ExternalReference: strconv.Itoa(pagamentoID), // <-- CORREÇÃO AQUI
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Erro ao fazer marshal do payload da cobrança Asaas: %v", err)
		return nil, fmt.Errorf("erro ao preparar dados da cobrança: %w", err)
	}

	req, err := http.NewRequest("POST", asaasBaseURL+"/payments", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Erro ao criar requisição HTTP para cobrança Asaas: %v", err)
		return nil, fmt.Errorf("erro ao criar requisição: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("access_token", apiKey)

	log.Printf("Tentando criar cobrança Asaas. URL: %s/payments, Payload: %s", asaasBaseURL, string(jsonData)) // <-- LOG ADICIONADO

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Erro ao executar client.Do para criar cobrança Asaas: %v", err) // <-- LOG ADICIONADO
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Log detalhado do erro da cobrança
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			log.Printf("Falha ao criar cobrança no Asaas (Status: %d) e falha ao ler corpo da resposta: %v", resp.StatusCode, readErr)
			return nil, errors.New("falha ao criar cobrança no Asaas e ler resposta")
		}
		log.Printf("Falha ao criar cobrança no Asaas (Status: %d). Resposta Asaas: %s", resp.StatusCode, string(bodyBytes))
		return nil, errors.New("falha ao criar cobrança no Asaas")
	}

	var chargeResponse AsaasChargeResponse
	if err := json.NewDecoder(resp.Body).Decode(&chargeResponse); err != nil {
		log.Printf("Erro ao decodificar resposta da criação de cobrança Asaas: %v", err)
		return nil, fmt.Errorf("erro ao processar resposta da cobrança: %w", err)
	}

	return &chargeResponse, nil
}

// AsaasCustomerRequest é para criar um novo cliente
type AsaasCustomerRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	CpfCnpj string `json:"cpfCnpj"`
}

// AsaasCustomerResponse é a resposta da API de clientes
type AsaasCustomerResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AsaasCustomerListResponse é para buscar clientes
type AsaasCustomerListResponse struct {
	Data []AsaasCustomerResponse `json:"data"`
}

// FindOrCreateCustomer busca um cliente pelo CPF/CNPJ ou o cria se não existir.
// Retorna o ID do cliente no Asaas.
func FindOrCreateCustomer(nome, email, cpf string) (string, error) {
	apiKey := os.Getenv("ASAAS_API_KEY")
	if apiKey == "" {
		return "", errors.New("ASAAS_API_KEY não encontrada")
	}

	client := &http.Client{Timeout: 10 * time.Second} // Definido uma vez

	// 1. Tenta buscar o cliente pelo CPF
	searchURL := asaasBaseURL + "/customers?cpfCnpj=" + cpf
	reqSearch, errSearch := http.NewRequest("GET", searchURL, nil)
	if errSearch != nil {
		log.Printf("Erro ao criar requisição HTTP para buscar cliente Asaas: %v", errSearch)
		return "", fmt.Errorf("erro ao criar requisição de busca: %w", errSearch)
	}
	reqSearch.Header.Set("access_token", apiKey)

	respSearch, errSearch := client.Do(reqSearch)
	if errSearch != nil {
		log.Printf("Erro ao executar client.Do para buscar cliente Asaas: %v", errSearch)
		return "", errSearch
	}
	defer respSearch.Body.Close()

	if respSearch.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(respSearch.Body)
		log.Printf("Falha ao buscar cliente no Asaas (Status: %d). Resposta Asaas: %s", respSearch.StatusCode, string(bodyBytes))
		return "", errors.New("falha ao buscar cliente no Asaas")
	}

	var searchResult AsaasCustomerListResponse
	if err := json.NewDecoder(respSearch.Body).Decode(&searchResult); err != nil {
		log.Printf("Erro ao decodificar resposta da busca de cliente Asaas: %v", err)
		return "", fmt.Errorf("erro ao processar resposta da busca: %w", err)
	}

	if len(searchResult.Data) > 0 {
		log.Printf("Cliente Asaas encontrado para CPF %s. ID: %s", cpf, searchResult.Data[0].ID)
		return searchResult.Data[0].ID, nil
	}

	log.Printf("Cliente Asaas não encontrado para CPF %s. Tentando criar...", cpf)

	// 2. Se não encontrou, cria um novo
	customerPayload := AsaasCustomerRequest{Name: nome, Email: email, CpfCnpj: cpf}
	jsonData, jsonErr := json.Marshal(customerPayload)
	if jsonErr != nil {
		log.Printf("Erro ao fazer marshal do payload do cliente Asaas: %v", jsonErr)
		return "", fmt.Errorf("erro ao preparar dados do cliente: %w", jsonErr)
	}

	createURL := asaasBaseURL + "/customers"
	reqCreate, reqErr := http.NewRequest("POST", createURL, bytes.NewBuffer(jsonData))
	if reqErr != nil {
		log.Printf("Erro ao criar requisição HTTP para criar cliente Asaas: %v", reqErr)
		return "", fmt.Errorf("erro ao criar requisição: %w", reqErr)
	}
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("access_token", apiKey)

	log.Printf("Tentando criar cliente Asaas. URL: %s, Payload: %s", createURL, string(jsonData))

	respCreate, errCreate := client.Do(reqCreate)
	if errCreate != nil {
		log.Printf("Erro ao executar client.Do para criar cliente Asaas: %v", errCreate)
		return "", errCreate
	}
	defer respCreate.Body.Close()

	// --- CORREÇÃO DO ESCOPO AQUI ---
	// Lê o corpo da resposta UMA VEZ
	bodyBytes, readErr := io.ReadAll(respCreate.Body)
	if readErr != nil {
		log.Printf("Falha ao criar cliente no Asaas (Status: %d) e falha ao ler corpo da resposta: %v", respCreate.StatusCode, readErr)
		return "", errors.New("falha ao criar cliente no Asaas e ler resposta")
	}

	// Verifica o Status Code da Criação
	if respCreate.StatusCode != http.StatusOK {
		// Log da resposta específica do Asaas (usa bodyBytes que lemos antes)
		log.Printf("Falha ao criar cliente no Asaas (Status: %d). Resposta Asaas: %s", respCreate.StatusCode, string(bodyBytes))
		return "", errors.New("falha ao criar cliente no Asaas")
	}
	// --- FIM DA CORREÇÃO DO ESCOPO ---

	// Se chegou aqui, StatusCode é OK (200)
	var newCustomer AsaasCustomerResponse
	// Usa Unmarshal pois já lemos o bodyBytes
	if err := json.Unmarshal(bodyBytes, &newCustomer); err != nil {
		log.Printf("Erro ao decodificar resposta da criação de cliente Asaas (Status OK): %v. Body: %s", err, string(bodyBytes))
		return "", fmt.Errorf("erro ao processar resposta da criação de cliente: %w", err)
	}

	log.Printf("Cliente Asaas criado com sucesso para CPF %s. ID: %s", cpf, newCustomer.ID)
	return newCustomer.ID, nil
}
