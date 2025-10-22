// internal/email/email.go
package email

import (
	"fmt"
	"log"
	"os"
	"strings" // Importe 'strings'

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// ReceiptData contém as informações para o recibo
type ReceiptData struct {
	ToEmail    string
	ToName     string
	RifaTitle  string
	Numbers    []int
	TotalValue float64
}

// SendConfirmationEmail envia o e-mail de confirmação da compra.
func SendConfirmationEmail(data ReceiptData) {
	apiKey := os.Getenv("SENDGRID_API_KEY")
	senderEmail := os.Getenv("SENDER_EMAIL")
	senderName := "Rifa Online" // Ou o nome do seu projeto

	if apiKey == "" || senderEmail == "" {
		log.Println("AVISO: SENDGRID_API_KEY or SENDER_EMAIL não estão definidos. E-mail não enviado.")
		return
	}

	from := mail.NewEmail(senderName, senderEmail)
	to := mail.NewEmail(data.ToName, data.ToEmail)

	// Formata a lista de números
	// Converte o slice []int para []string
	var numbersStr []string
	for _, n := range data.Numbers {
		numbersStr = append(numbersStr, fmt.Sprintf("%d", n))
	}
	// Junta com vírgula: "1, 2, 3"
	numbersList := strings.Join(numbersStr, ", ")

	// --- Conteúdo do E-mail (HTML) ---
	// Você pode deixar este HTML muito mais bonito depois
	htmlContent := fmt.Sprintf(`
		<html>
		<body>
			<h1>Pagamento Confirmado!</h1>
			<p>Olá, <strong>%s</strong>!</p>
			<p>Seu pagamento para a rifa "<strong>%s</strong>" foi confirmado com sucesso.</p>
			<p>Boa sorte! Seus números são:</p>
			<h2 style="font-size: 24px; background-color: #f0f0f0; padding: 10px; border-radius: 5px;">
				%s
			</h2>
			<p>Valor total: R$ %.2f</p>
			<br/>
			<p>Obrigado por participar!</p>
		</body>
		</html>
	`, data.ToName, data.RifaTitle, numbersList, data.TotalValue)

	// --- Fim do Conteúdo ---

	plainTextContent := "Seu pagamento foi confirmado! Seus números são: " + numbersList
	subject := "Confirmação de Pagamento da Rifa: " + data.RifaTitle

	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	client := sendgrid.NewSendClient(apiKey)

	response, err := client.Send(message)
	if err != nil {
		log.Printf("Erro ao enviar e-mail de confirmação para %s: %v", data.ToEmail, err)
	} else if response.StatusCode >= 400 {
		// Log de erro do SendGrid
		log.Printf("Erro do SendGrid ao enviar para %s (Status: %d): %s", data.ToEmail, response.StatusCode, response.Body)
	} else {
		log.Printf("E-mail de confirmação enviado com sucesso para %s (Status: %d)", data.ToEmail, response.StatusCode)
	}
}
