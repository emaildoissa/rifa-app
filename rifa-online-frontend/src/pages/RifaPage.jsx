// src/pages/RifaPage.jsx

import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import api from '../services/api';
import styles from './RifaPage.module.css';

// --- CONFIGURAÇÃO PIX MANUAL ---
const MINHA_CHAVE_PIX = "seu-email@exemplo.com"; // <-- TROQUE AQUI
const MEU_WHATSAPP = "(99) 99999-9999"; // <-- TROQUE AQUI
// --- FIM DA CONFIGURAÇÃO ---

function RifaPage() {
  const { id } = useParams();
  const [rifa, setRifa] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedNumbers, setSelectedNumbers] = useState([]);
  const [formData, setFormData] = useState({
    nome_comprador: '',
    email_comprador: '',
    telefone_comprador: '',
    cpf_cnpj: '',
  });
  const [isReserving, setIsReserving] = useState(false);
  const [reservationError, setReservationError] = useState(null);
  const [reservationSuccessData, setReservationSuccessData] = useState(null);
  const [copySuccess, setCopySuccess] = useState('');
  const [currentStep, setCurrentStep] = useState('selecting'); // 'selecting' ou 'fillingForm'

  useEffect(() => {
    const fetchRifaDetails = async () => {
      try {
        const response = await api.get(`/rifas/${id}`);
        setRifa(response.data);
      } catch (err) {
        setError('Não foi possível carregar os detalhes desta rifa.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchRifaDetails();
  }, [id]);

  const handleNumberClick = (number) => {
    if (currentStep !== 'selecting') return;
    if (selectedNumbers.includes(number)) {
      setSelectedNumbers(selectedNumbers.filter((n) => n !== number));
    } else {
      setSelectedNumbers([...selectedNumbers, number]);
    }
  };

  const handleFormChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };

  const handleReservation = async (e) => {
    e.preventDefault();
    setIsReserving(true);
    setReservationError(null);
    const payload = { ...formData, numeros: selectedNumbers };
    try {
      const response = await api.post(`/rifas/${id}/reservar`, payload);
      setReservationSuccessData(response.data);
    } catch (err) {
      let errorMsg = 'Erro ao processar sua reserva. Tente novamente.';
      if (err.response?.data?.error) {
        errorMsg = err.response.data.error;
      }
      setReservationError(errorMsg);
      console.error(err);
    } finally {
      setIsReserving(false);
    }
  };

  const getButtonClassName = (number) => {
    let baseStyle = styles.numberButton;
    if (number.status !== 'disponivel') {
      return `${baseStyle} ${styles.indisponivel}`;
    }
    if (selectedNumbers.includes(number.numero)) {
      return `${baseStyle} ${styles.selecionado}`;
    }
    const interactionClass = currentStep !== 'selecting' ? styles.disabledInteraction : '';
    return `${baseStyle} ${styles.disponivel} ${interactionClass}`;
  };

  const handleCopyPixKey = () => {
    navigator.clipboard.writeText(MINHA_CHAVE_PIX).then(() => {
      setCopySuccess('Chave PIX copiada!');
      setTimeout(() => setCopySuccess(''), 2000);
    }, (err) => {
      setCopySuccess('Falha ao copiar.');
      console.error('Falha ao copiar PIX: ', err);
    });
  };

  const proceedToForm = () => setCurrentStep('fillingForm');
  const backToSelection = () => setCurrentStep('selecting');

  // --- RENDERIZAÇÃO ---
  if (loading) return <div className="text-center p-10">Carregando detalhes da rifa...</div>;
  if (error) return <div className="text-center p-10 text-red-500">Erro: {error}</div>;
  if (!rifa) return <div>Rifa não encontrada.</div>;

  if (reservationSuccessData) {
    // Tela de Sucesso (PIX Manual)
    return (
      <div className={styles.successContainer}>
        <h2 className={styles.successTitle}>Reserva Realizada!</h2>
        <p>Seus números foram reservados com sucesso. Para confirmar, realize o pagamento via PIX e envie o comprovante.</p>
        <p className={styles.summaryTotal}>
          <strong>Valor Total:</strong> R$ {reservationSuccessData.valor.toFixed(2)}
        </p>
        <h3>Pague com esta Chave PIX:</h3>
        <div className={styles.pixKeyBox}>
          <span className={styles.pixKey}>{MINHA_CHAVE_PIX}</span>
          <button onClick={handleCopyPixKey} className={styles.copyButton}>
            {copySuccess ? copySuccess : 'Copiar Chave'}
          </button>
        </div>
        <p className={styles.pixInstructions}>
          <strong>Importante:</strong> Envie o comprovante de pagamento para o WhatsApp: <strong> {MEU_WHATSAPP} </strong> para validarmos sua compra.
        </p>
        <p className={styles.thankYou}>Obrigado por participar e boa sorte!</p>
      </div>
    );
  }

  // Página Principal (Seleção ou Formulário)
  return (
    <div className={styles.container}>
      {/* Detalhes da Rifa */}
      <div className={styles.rifaHeader}>
        <h1 className={styles.rifaTitle}>{rifa.titulo}</h1>
        <p className={styles.rifaPrize}><strong>Prêmio:</strong> {rifa.premio}</p>
        <p className={styles.rifaPrice}>R$ {rifa.preco_por_numero.toFixed(2)}</p>
        <p className={styles.rifaDescription}>{rifa.descricao}</p>
      </div>

      {/* Formulário (só aparece na etapa 'fillingForm') */}
      {currentStep === 'fillingForm' && (
        <div className={styles.reservationBox}>
          <button onClick={backToSelection} className={styles.backButton}>
            &larr; Voltar à Seleção de Números
          </button>
          <h3 className={styles.reservationTitle}>Finalizar Reserva</h3>
          <p className={styles.summaryText}>
            <strong>Números:</strong> {selectedNumbers.sort((a, b) => a - b).join(', ')}
          </p>
          <p className={styles.summaryText}>
            <strong>Quantidade:</strong> {selectedNumbers.length}
          </p>
          <p className={styles.summaryTotal}>
            <strong>Total:</strong> R$ {(selectedNumbers.length * rifa.preco_por_numero).toFixed(2)}
          </p>
          <form onSubmit={handleReservation} style={{ marginTop: '20px' }}>
            <div className={styles.formGroup}>
              <label htmlFor="nome_comprador">Nome Completo:</label>
              <input type="text" id="nome_comprador" name="nome_comprador" onChange={handleFormChange} required className={styles.formInput} />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="email_comprador">Email:</label>
              <input type="email" id="email_comprador" name="email_comprador" onChange={handleFormChange} required className={styles.formInput} />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="telefone_comprador">Telefone (com DDD):</label>
              <input type="tel" id="telefone_comprador" name="telefone_comprador" onChange={handleFormChange} required className={styles.formInput} />
            </div>
            <div className={styles.formGroup}>
              <label htmlFor="cpf_cnpj">CPF (apenas números):</label>
              <input type="text" id="cpf_cnpj" name="cpf_cnpj" onChange={handleFormChange} required className={styles.formInput} />
            </div>
            <button type="submit" disabled={isReserving} className={styles.submitButton}>
              {isReserving ? 'Reservando...' : 'Confirmar Reserva e Ver Instruções PIX'}
            </button>
            {reservationError && (
              <p className={styles.errorMessage}><strong>Erro:</strong> {reservationError}</p>
            )}
          </form>
        </div>
      )}

      {/* --- Grade de Números (só aparece na etapa 'selecting') --- */}
      {currentStep === 'selecting' && (
        <> {/* Fragmento para agrupar o título e a grade */}
          <h2 className={styles.numbersTitle}>Escolha seus números:</h2>
          <div className={styles.numbersGrid}>
            {rifa.numeros.map((numero) => (
              <button
                key={numero.id}
                className={getButtonClassName(numero)}
                disabled={numero.status !== 'disponivel'} // Só desabilita se não disponível
                onClick={() => handleNumberClick(numero.numero)}
              >
                {numero.numero}
              </button>
            ))}
          </div>
        </>
      )}

      {/* Botão para Proceder ao Formulário (só aparece na etapa 'selecting' e se houver números) */}
      {currentStep === 'selecting' && selectedNumbers.length > 0 && (
        <div className={styles.proceedButtonContainer}>
           <p>
            {selectedNumbers.length} número(s) selecionado(s) - Total: R$ {(selectedNumbers.length * rifa.preco_por_numero).toFixed(2)}
          </p>
          <button onClick={proceedToForm} className={styles.proceedButton}>
            Confirmar Números e Continuar
          </button>
        </div>
      )}

    </div>
  );
}

export default RifaPage;