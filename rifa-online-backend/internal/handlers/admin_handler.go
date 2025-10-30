// internal/handlers/admin_handler.go
package handlers

import (
	"context"
	"log"
	"net/http"
	"rifa-online-backend/internal/database"
	"rifa-online-backend/internal/email" // Importar email
	"rifa-online-backend/internal/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPendingPagamentos lista todos os pagamentos pendentes para o dashboard
func GetPendingPagamentos(c *gin.Context) {
	var pagamentos []models.PagamentoPendente

	// Query complexa para juntar tudo que o admin precisa ver
	query := `
		SELECT 
			p.id AS pagamento_id,
			r.id AS rifa_id,
			r.titulo AS rifa_titulo,
			p.valor_total,
			n.nome_comprador,
			n.email_comprador,
			n.telefone_comprador,
			MIN(n.data_reserva) AS data_reserva,
			ARRAY_AGG(n.numero ORDER BY n.numero) AS numeros
		FROM pagamentos p
		JOIN pagamento_numeros pn ON p.id = pn.pagamento_id
		JOIN numeros n ON pn.numero_id = n.id
		JOIN rifas r ON n.rifa_id = r.id
		WHERE p.status = 'pendente'
		GROUP BY p.id, r.id, r.titulo, p.valor_total, n.nome_comprador, n.email_comprador, n.telefone_comprador
		ORDER BY p.id ASC; -- <-- CORREÇÃO AQUI
	`

	rows, err := database.DB.Query(context.Background(), query)
	if err != nil {
		log.Printf("Erro ao buscar pagamentos pendentes: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar pagamentos"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var pg models.PagamentoPendente
		var pgNumbers []int32 // ARRAY_AGG de int vem como []int32
		if err := rows.Scan(
			&pg.PagamentoID,
			&pg.RifaID,
			&pg.RifaTitulo,
			&pg.ValorTotal,
			&pg.NomeComprador,
			&pg.EmailComprador,
			&pg.TelefoneComprador,
			&pg.DataReserva,
			&pgNumbers,
		); err != nil {
			log.Printf("Erro ao escanear pagamento pendente: %v", err)
			continue
		}

		// Converter []int32 para []int
		pg.Numeros = make([]int, len(pgNumbers))
		for i, v := range pgNumbers {
			pg.Numeros[i] = int(v)
		}

		pagamentos = append(pagamentos, pg)
	}

	if pagamentos == nil {
		pagamentos = make([]models.PagamentoPendente, 0)
	}

	c.JSON(http.StatusOK, pagamentos)
}

// AprovarPagamento aprova um pagamento manual
func AprovarPagamento(c *gin.Context) {
	pagamentoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pagamento inválido"})
		return
	}

	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar transação"})
		return
	}
	defer tx.Rollback(context.Background())

	// 1. Atualizar o status do pagamento
	tag, err := tx.Exec(context.Background(),
		"UPDATE pagamentos SET status = 'aprovado' WHERE id = $1 AND status = 'pendente'",
		pagamentoID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar pagamento"})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado ou já foi aprovado"})
		return
	}

	// 2. Atualizar o status dos números
	_, err = tx.Exec(context.Background(), `
		UPDATE numeros SET status = 'pago'
		WHERE id IN (SELECT numero_id FROM pagamento_numeros WHERE pagamento_id = $1)
	`, pagamentoID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao atualizar status dos números"})
		return
	}

	// 3. Comitar a transação ANTES de enviar o e-mail
	if err := tx.Commit(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao comitar transação"})
		return
	}

	// 4. Disparar e-mail de confirmação (em background)
	// Usamos a mesma lógica do antigo webhook
	go func(pID int) {
		var data email.ReceiptData
		var pgNumbers []int32

		query := `
			SELECT 
				n.nome_comprador, 
				n.email_comprador, 
				r.titulo, 
				p.valor_total,
				ARRAY_AGG(n.numero ORDER BY n.numero) as numeros
			FROM pagamentos p
			JOIN pagamento_numeros pn ON p.id = pn.pagamento_id
			JOIN numeros n ON pn.numero_id = n.id
			JOIN rifas r ON n.rifa_id = r.id
			WHERE p.id = $1
			GROUP BY n.nome_comprador, n.email_comprador, r.titulo, p.valor_total
		`
		// Note que adicionamos p.valor_total ao SELECT e GROUP BY
		err := database.DB.QueryRow(context.Background(), query, pID).Scan(
			&data.ToName,
			&data.ToEmail,
			&data.RifaTitle,
			&data.TotalValue, // Pegamos o valor do banco
			&pgNumbers,
		)

		if err != nil {
			log.Printf("ERRO (EMAIL): Não foi possível buscar dados do recibo para pagamento %d: %v", pID, err)
			return
		}

		data.Numbers = make([]int, len(pgNumbers))
		for i, v := range pgNumbers {
			data.Numbers[i] = int(v)
		}

		email.SendConfirmationEmail(data)
	}(pagamentoID)

	c.JSON(http.StatusOK, gin.H{"message": "Pagamento aprovado com sucesso! E-mail de confirmação enviado."})
}

// LiberarPagamento (Cancela a reserva)
func LiberarPagamento(c *gin.Context) {
	pagamentoID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID de pagamento inválido"})
		return
	}

	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao iniciar transação"})
		return
	}
	defer tx.Rollback(context.Background())

	// 1. Reverter o status dos números para 'disponivel'
	// Também limpamos os dados do comprador e a data da reserva
	_, err = tx.Exec(context.Background(), `
		UPDATE numeros 
		SET status = 'disponivel', nome_comprador = NULL, email_comprador = NULL, telefone_comprador = NULL, data_reserva = NULL
		WHERE id IN (
			SELECT numero_id FROM pagamento_numeros 
			WHERE pagamento_id = $1
		)
	`, pagamentoID)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao liberar os números"})
		return
	}

	// 2. Mudar o status do pagamento para 'cancelado' (ou deletar)
	// Mudar o status é melhor para manter o histórico.
	tag, err := tx.Exec(context.Background(),
		"UPDATE pagamentos SET status = 'cancelado' WHERE id = $1 AND status = 'pendente'",
		pagamentoID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao cancelar o pagamento"})
		return
	}

	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pagamento não encontrado ou não está mais pendente"})
		return
	}

	// 3. Comitar a transação
	if err := tx.Commit(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao comitar transação"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reserva cancelada e números liberados."})
}
func GetParticipantesPorRifa(c *gin.Context) {
	rifaID, err := strconv.Atoi(c.Param("id")) // Pega o ID da rifa da URL
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da rifa inválido"})
		return
	}

	var participantes []models.ParticipanteInfo

	// Query para buscar números pagos, agrupados por comprador
	query := `
		SELECT 
			nome_comprador,
			email_comprador,
			telefone_comprador,
			ARRAY_AGG(numero ORDER BY numero) AS numeros
		FROM numeros
		WHERE rifa_id = $1 AND status = 'pago' -- Somente números com status 'pago'
		GROUP BY nome_comprador, email_comprador, telefone_comprador -- Agrupa por pessoa
		ORDER BY nome_comprador ASC; -- Ordena por nome
	`

	rows, err := database.DB.Query(context.Background(), query, rifaID)
	if err != nil {
		log.Printf("Erro ao buscar participantes da rifa %d: %v", rifaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar participantes"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var p models.ParticipanteInfo
		var pgNumbers []int32 // ARRAY_AGG vem como []int32
		if err := rows.Scan(
			&p.NomeComprador,
			&p.EmailComprador,
			&p.TelefoneComprador,
			&pgNumbers,
		); err != nil {
			log.Printf("Erro ao escanear participante: %v", err)
			continue
		}

		// Converter []int32 para []int
		p.Numeros = make([]int, len(pgNumbers))
		for i, v := range pgNumbers {
			p.Numeros[i] = int(v)
		}

		participantes = append(participantes, p)
	}

	if participantes == nil {
		participantes = make([]models.ParticipanteInfo, 0)
	}

	c.JSON(http.StatusOK, participantes)
}
