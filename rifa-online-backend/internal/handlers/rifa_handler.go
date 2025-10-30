package handlers

import (
	"context"
	"log"
	"net/http"
	"rifa-online-backend/internal/database"
	"rifa-online-backend/internal/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type RifaDetail struct {
	models.Rifa
	Numeros []models.Numero `json:"numeros"`
	Winner  *models.Numero  `json:"winner,omitempty"`
}

// CreateRifa é o handler para criar uma nova rifa e seus números associados.
func CreateRifa(c *gin.Context) {
	var input models.Rifa

	// Valida o JSON de entrada com base nas tags 'binding' da struct
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Inicia uma transação com o banco de dados
	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		log.Printf("Erro ao iniciar a transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}
	// Garante que a transação será desfeita (rollback) se algo der errado
	defer tx.Rollback(context.Background())

	// 1. Inserir a rifa na tabela 'rifas' e obter o ID gerado
	sqlRifa := `
		INSERT INTO rifas (titulo, descricao, premio, preco_por_numero, total_numeros, data_sorteio, status, imagem_url)
		VALUES ($1, $2, $3, $4, $5, $6, 'ativa')
		RETURNING id, created_at, updated_at, status
	`
	err = tx.QueryRow(context.Background(), sqlRifa,
		input.Titulo,         // $1
		input.Descricao,      // $2
		input.Premio,         // $3
		input.PrecoPorNumero, // $4
		input.TotalNumeros,   // $5
		input.DataSorteio,    // $6
		"ativa",              // $7 (para status)
		input.ImagemURL,      // $8 (agora é *string, pgx lida com nil)
	).Scan(&input.ID, &input.CreatedAt, &input.UpdatedAt, &input.Status)

	if err != nil {
		log.Printf("Erro ao inserir rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar a rifa"})
		return
	}

	// 2. Gerar os números da rifa em lote (batch insert)
	numerosParaInserir := make([][]interface{}, input.TotalNumeros)
	for i := 0; i < input.TotalNumeros; i++ {
		numerosParaInserir[i] = []interface{}{input.ID, i + 1}
	}

	_, err = tx.CopyFrom(
		context.Background(),
		pgx.Identifier{"numeros"},
		[]string{"rifa_id", "numero"},
		pgx.CopyFromRows(numerosParaInserir),
	)

	if err != nil {
		log.Printf("Erro ao inserir números da rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao gerar os números da rifa"})
		return
	}

	// Se tudo correu bem, comita a transação
	if err := tx.Commit(context.Background()); err != nil {
		log.Printf("Erro ao comitar a transação: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao finalizar a criação da rifa"})
		return
	}

	// Retorna a rifa recém-criada como resposta
	c.JSON(http.StatusCreated, input)
}

func GetRifaByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}
	var rifa models.Rifa
	var numeros []models.Numero

	// 1. Busca os dados principais da rifa (incluindo numero_sorteado)
	sqlRifa := `SELECT id, titulo, descricao, premio, preco_por_numero, total_numeros, data_sorteio, status, imagem_url, numero_sorteado FROM rifas WHERE id = $1`
	err = database.DB.QueryRow(context.Background(), sqlRifa, id).Scan(
		&rifa.ID, &rifa.Titulo, &rifa.Descricao, &rifa.Premio, &rifa.PrecoPorNumero, &rifa.TotalNumeros, &rifa.DataSorteio, &rifa.Status, &rifa.ImagemURL,
		&rifa.NumeroSorteado, // Scan do número sorteado
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rifa não encontrada"})
			return
		}
		log.Printf("Erro ao buscar rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar dados da rifa"})
		return
	}

	// 2. Busca os dados do Ganhador (se a rifa estiver sorteada)
	var winnerData models.Numero
	if rifa.Status == "sorteada" && rifa.NumeroSorteado > 0 {
		sqlWinner := `
            SELECT id, rifa_id, numero, status, nome_comprador, email_comprador, telefone_comprador, data_reserva
            FROM numeros 
            WHERE rifa_id = $1 AND numero = $2
        `
		// Note que o modelo Numero usa ponteiros, lidando com NULLs
		errWinner := database.DB.QueryRow(context.Background(), sqlWinner, id, rifa.NumeroSorteado).Scan(
			&winnerData.ID, &winnerData.RifaID, &winnerData.Numero, &winnerData.Status,
			&winnerData.NomeComprador, &winnerData.EmailComprador, &winnerData.TelefoneComprador, &winnerData.DataReserva,
		)
		if errWinner != nil && errWinner != pgx.ErrNoRows {
			// Se der erro (exceto 'não encontrado'), apenas loga, não quebra a requisição
			log.Printf("Aviso: Rifa %d está sorteada (número %d), mas não foi possível encontrar dados do ganhador: %v", id, rifa.NumeroSorteado, errWinner)
		}
	}

	// 3. Busca todos os números (como antes)
	sqlNumeros := `
    SELECT id, rifa_id, numero, status, nome_comprador, email_comprador, telefone_comprador
    FROM numeros
    WHERE rifa_id = $1
    ORDER BY numero ASC
`
	rows, err := database.DB.Query(context.Background(), sqlNumeros, id)
	if err != nil {
		log.Printf("Erro ao buscar números da rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar os números da rifa"})
		return
	}
	defer rows.Close()
	for rows.Next() {
		var n models.Numero
		if err := rows.Scan(&n.ID, &n.RifaID, &n.Numero, &n.Status, &n.NomeComprador, &n.EmailComprador, &n.TelefoneComprador); err != nil {
			log.Printf("Erro ao escanear linha de número: %v", err)
			continue
		}
		numeros = append(numeros, n)
	}

	// 4. Monta a resposta final
	response := RifaDetail{
		Rifa:    rifa,
		Numeros: numeros,
	}

	// Adiciona o ganhador à resposta se ele foi encontrado
	if winnerData.ID != 0 {
		response.Winner = &winnerData
	}

	c.JSON(http.StatusOK, response)
}

func GetAllRifas(c *gin.Context) {
	var rifas []models.RifaSummary

	sql := `
		SELECT id, titulo, premio, preco_por_numero, status, imagem_url
		FROM rifas
		WHERE status = 'ativa'
		ORDER BY created_at DESC
	`

	rows, err := database.DB.Query(context.Background(), sql)
	if err != nil {
		log.Printf("Erro ao buscar rifas: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar rifas"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RifaSummary
		if err := rows.Scan(&r.ID, &r.Titulo, &r.Premio, &r.PrecoPorNumero, &r.Status, &r.ImagemURL); err != nil {
			log.Printf("Erro ao escanear linha da rifa: %v", err)
			continue
		}
		rifas = append(rifas, r)
	}

	// É uma boa prática retornar um array vazio `[]` em vez de `null` se não houver resultados
	if rifas == nil {
		rifas = make([]models.RifaSummary, 0)
	}

	c.JSON(http.StatusOK, rifas)
}

func ReservarNumeros(c *gin.Context) {
	idStr := c.Param("id")
	rifaID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da rifa inválido"})
		return
	}
	var input models.ReservaInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tx, err := database.DB.Begin(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro interno do servidor"})
		return
	}
	defer tx.Rollback(context.Background())
	var rifa models.Rifa
	err = tx.QueryRow(context.Background(), "SELECT preco_por_numero, titulo, status FROM rifas WHERE id = $1", rifaID).Scan(&rifa.PrecoPorNumero, &rifa.Titulo, &rifa.Status)
	if err != nil || rifa.Status != "ativa" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rifa não encontrada ou não está ativa"})
		return
	}

	// --- 3. CORREÇÃO AQUI ---
	const tempoExpiracao = "15 minutes"
	// Mesma lógica de usar fmt.Sprintf
	/* queryUpdate := fmt.Sprintf(`
	        UPDATE numeros
	        SET
	            status = 'reservado',
	            nome_comprador = $1,
	            email_comprador = $2,
	            telefone_comprador = $3,
	            data_reserva = $4
	        WHERE
	            rifa_id = $5 AND
	            numero = ANY($6::int[]) AND
	            (status = 'disponivel' OR (status = 'reservado' AND data_reserva < (NOW() - INTERVAL '%s')))
	        RETURNING id
	    `, tempoExpiracao)

		// Agora a chamada tem 6 argumentos (o $7 foi removido da chamada)
		rows, err := tx.Query(context.Background(), queryUpdate,
			input.NomeComprador,
			input.EmailComprador,
			input.TelefoneComprador,
			time.Now(),    // $4
			rifaID,        // $5
			input.Numeros, // $6
			// 'tempoExpiracao' não é mais passado como argumento
		) */
	queryUpdate := `
    UPDATE numeros 
    SET
        status = 'reservado',
        nome_comprador = $1,
        email_comprador = $2,
        telefone_comprador = $3,
        data_reserva = $4
    WHERE
        rifa_id = $5 AND
        numero = ANY($6::int[]) AND
        status = 'disponivel' -- Somente permite reservar se estiver disponível
    RETURNING id
`
	// A chamada volta a ter 6 argumentos
	rows, err := tx.Query(context.Background(), queryUpdate,
		input.NomeComprador,
		input.EmailComprador,
		input.TelefoneComprador,
		time.Now(),    // $4
		rifaID,        // $5
		input.Numeros, // $6
	)
	// --- FIM DA CORREÇÃO ---

	if err != nil {
		log.Printf("Erro ao reservar números: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao executar a reserva"})
		return
	}
	var reservedNumberIDs []int
	for rows.Next() {
		var id int
		rows.Scan(&id)
		reservedNumberIDs = append(reservedNumberIDs, id)
	}
	if len(reservedNumberIDs) != len(input.Numeros) {
		c.JSON(http.StatusConflict, gin.H{"error": "Um ou mais números não estão disponíveis"})
		return
	}
	rows.Close()
	valorTotal := rifa.PrecoPorNumero * float64(len(input.Numeros))
	var pagamentoID int
	sqlPagamento := `INSERT INTO pagamentos (status, valor_total, id_transacao_gateway) VALUES ('pendente', $1, '') RETURNING id`
	err = tx.QueryRow(context.Background(), sqlPagamento, valorTotal).Scan(&pagamentoID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao criar registro de pagamento"})
		return
	}
	for _, numeroID := range reservedNumberIDs {
		_, err = tx.Exec(context.Background(), "INSERT INTO pagamento_numeros (pagamento_id, numero_id) VALUES ($1, $2)", pagamentoID, numeroID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao vincular pagamento aos números"})
			return
		}
	}
	/* customerID, err := payment.FindOrCreateCustomer(input.NomeComprador, input.EmailComprador, input.CpfCnpj)
	if err != nil {
		log.Printf("Erro no Asaas (cliente): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro no provedor de pagamento"})
		return
	}
	chargeResponse, err := payment.CreatePixCharge(customerID, rifa, valorTotal, pagamentoID)
	if err != nil {
		log.Printf("Erro no Asaas (cobrança): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao gerar cobrança PIX"})
		return
	}
	_, err = tx.Exec(context.Background(), "UPDATE pagamentos SET id_transacao_gateway = $1 WHERE id = $2", chargeResponse.ID, pagamentoID)*/
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar ID do gateway"})
		return
	}
	if err := tx.Commit(context.Background()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao finalizar a reserva"})
		return
	}
	//c.JSON(http.StatusOK, chargeResponse)
	// --- NOVA RESPOSTA ---
	c.JSON(http.StatusOK, gin.H{
		"message":   "Reserva criada com sucesso! Envie o comprovante.",
		"paymentId": pagamentoID, // Envia o ID para o frontend
		"valor":     valorTotal,
	})
}

func UpdateRifa(c *gin.Context) {
	// 1. Obter o ID da URL
	idStr := c.Param("id")
	rifaID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da rifa inválido"})
		return
	}

	// 2. Validar o corpo da requisição
	var input models.RifaUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Executar a atualização no banco de dados
	sql := `
		UPDATE rifas
		SET 
			titulo = $1,
			descricao = $2,
			premio = $3,
			data_sorteio = $4,
			status = $5,
			imagem_url = $6,
			updated_at = NOW()
		WHERE id = $7
	`
	tag, err := database.DB.Exec(context.Background(), sql,
		input.Titulo,
		input.Descricao,
		input.Premio,
		input.DataSorteio,
		input.Status,
		input.ImagemURL,
		rifaID,
	)

	if err != nil {
		log.Printf("Erro ao atualizar rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao atualizar rifa"})
		return
	}

	// 4. Verificar se alguma linha foi realmente atualizada
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rifa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rifa atualizada com sucesso"})
}

// DeleteRifa é o handler para apagar uma rifa.
func DeleteRifa(c *gin.Context) {
	// 1. Obter o ID da URL
	idStr := c.Param("id")
	rifaID, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID da rifa inválido"})
		return
	}

	// 2. Executar o DELETE
	// Graças ao "ON DELETE CASCADE" que definimos no SQL,
	// todos os 'numeros', 'pagamentos' e 'pagamento_numeros'
	// associados a esta rifa serão apagados automaticamente.
	sql := `DELETE FROM rifas WHERE id = $1`
	tag, err := database.DB.Exec(context.Background(), sql, rifaID)

	if err != nil {
		log.Printf("Erro ao deletar rifa: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Falha ao deletar rifa"})
		return
	}

	// 4. Verificar se alguma linha foi realmente deletada
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rifa não encontrada"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rifa apagada com sucesso"})
}

func GetAdminAllRifas(c *gin.Context) {
	var rifas []models.RifaSummary

	// Query modificada para incluir total_numeros e a contagem de vendidos (status='pago')
	sql := `
        SELECT 
            r.id, 
            r.titulo, 
            r.premio, 
            r.preco_por_numero, 
            r.status,
            r.total_numeros, 
			r.imagem_url,
            (SELECT COUNT(*) FROM numeros n WHERE n.rifa_id = r.id AND n.status = 'pago') AS numeros_vendidos
        FROM rifas r
        ORDER BY r.id DESC
    `

	rows, err := database.DB.Query(context.Background(), sql)
	if err != nil {
		log.Printf("Erro ao buscar rifas (admin): %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Erro ao buscar rifas"})
		return
	}
	defer rows.Close()

	for rows.Next() {
		var r models.RifaSummary

		if err := rows.Scan(
			&r.ID,
			&r.Titulo,
			&r.Premio,
			&r.PrecoPorNumero,
			&r.Status,
			&r.TotalNumeros,
			&r.ImagemURL,
			&r.NumerosVendidos,
		); err != nil {
			log.Printf("Erro ao escanear linha da rifa (admin): %v", err)
			continue
		}
		rifas = append(rifas, r)
	}

	if rifas == nil {
		rifas = make([]models.RifaSummary, 0)
	}

	c.JSON(http.StatusOK, rifas)
}
