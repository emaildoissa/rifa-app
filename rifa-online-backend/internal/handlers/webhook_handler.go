// internal/handlers/webhook_handler.go
package handlers

/* 	"context"
"log"
"net/http"
"os"
"rifa-online-backend/internal/database"
"rifa-online-backend/internal/email"
"rifa-online-backend/internal/models"

"github.com/gin-gonic/gin"
"github.com/jackc/pgx/v5" */

/* func AsaasWebhookHandler(c *gin.Context) {
	// 1. Validação de Segurança
	receivedToken := c.GetHeader("asaas-access-token")
	expectedToken := os.Getenv("ASAAS_WEBHOOK_TOKEN")

	if receivedToken != expectedToken {
		log.Println("AVISO: Tentativa de webhook com token inválido.")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token inválido"})
		return
	}

	// 2. Analisar o payload
	var payload models.AsaasWebhookPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		log.Printf("Erro ao decodificar payload do webhook: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Payload inválido"})
		return
	}

	log.Printf("Webhook recebido: Evento '%s', Pagamento ID '%s', Status '%s'", payload.Event, payload.Payment.ID, payload.Payment.Status)

	// 3. Processar apenas eventos de pagamento recebido
	if payload.Event != "PAYMENT_RECEIVED" {
		c.JSON(http.StatusOK, gin.H{"message": "Evento não processado"})
		return
	}

	// 4. Lógica de atualização no banco
	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		log.Printf("Erro ao iniciar transação do webhook: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	var pagamentoID int
	var valorTotal float64 // Precisamos do valor para o e-mail
	err = tx.QueryRow(context.Background(),
		"SELECT id, valor_total FROM pagamentos WHERE id_transacao_gateway = $1 AND status = 'pendente'", // Só atualiza se estiver pendente
		payload.Payment.ID,
	).Scan(&pagamentoID, &valorTotal)

	if err != nil {
		if err == pgx.ErrNoRows {
			log.Printf("Pagamento com ID de gateway %s não encontrado ou já processado.", payload.Payment.ID)
			c.JSON(http.StatusOK, gin.H{"error": "Pagamento não encontrado ou já processado"}) // Responde 200 para o Asaas não tentar de novo
			return
		}
		log.Printf("Erro: Pagamento com ID de gateway %s não encontrado no banco: %v", payload.Payment.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado"})
		return
	}

	// Atualizar o status do pagamento
	_, err = tx.Exec(context.Background(), "UPDATE pagamentos SET status = 'aprovado' WHERE id = $1", pagamentoID)
	if err != nil {
		log.Printf("Erro ao atualizar status do pagamento %d: %v", pagamentoID, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// Atualizar o status dos números
	_, err = tx.Exec(context.Background(), `
		UPDATE numeros SET status = 'pago'
		WHERE id IN (SELECT numero_id FROM pagamento_numeros WHERE pagamento_id = $1)
	`, pagamentoID)

	if err != nil {
		log.Printf("Erro ao atualizar status dos números para o pagamento %d: %v", pagamentoID, err)
		c.Status(http.StatusInternalServerError)
		return
	}

	// 5. Comitar a transação ANTES de enviar o e-mail
	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("Erro ao comitar transação do webhook: %v", err)
		c.Status(http.StatusInternalServerError)
		return
	}

	log.Printf("Pagamento %d (Asaas: %s) processado com sucesso. Números atualizados para 'pago'.", pagamentoID, payload.Payment.ID)

	// --- 6. NOVA LÓGICA DE E-MAIL ---
	// Dispara o e-mail em uma goroutine para não bloquear a resposta ao Asaas.
	go func(pID int, vTotal float64) {
		var data email.ReceiptData
		var pgNumbers []int32 // Use []int32 para o ARRAY_AGG de PostgreSQL

		// Query para buscar todos os dados necessários para o recibo
		query := `
			SELECT
				n.nome_comprador,
				n.email_comprador,
				r.titulo,
				ARRAY_AGG(n.numero ORDER BY n.numero) as numeros
			FROM pagamentos p
			JOIN pagamento_numeros pn ON p.id = pn.pagamento_id
			JOIN numeros n ON pn.numero_id = n.id
			JOIN rifas r ON n.rifa_id = r.id
			WHERE p.id = $1
			GROUP BY n.nome_comprador, n.email_comprador, r.titulo
		`
		err := database.DB.QueryRow(context.Background(), query, pID).Scan(
			&data.ToName,
			&data.ToEmail,
			&data.RifaTitle,
			&pgNumbers, // Escaneia o array de números
		)

		if err != nil {
			log.Printf("ERRO FATAL (EMAIL): Não foi possível buscar dados do recibo para pagamento %d: %v", pID, err)
			return // Não podemos enviar o e-mail
		}

		// Converte []int32 para []int
		data.Numbers = make([]int, len(pgNumbers))
		for i, v := range pgNumbers {
			data.Numbers[i] = int(v)
		}

		data.TotalValue = vTotal

		// Envia o e-mail
		email.SendConfirmationEmail(data)

	}(pagamentoID, valorTotal) // Passa as variáveis para a goroutine

	// 7. Responder ao Asaas que tudo foi recebido com sucesso
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
*/
