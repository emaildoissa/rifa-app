// src/pages/RifaPage.jsx

import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import api from '../services/api';

// 1. Importe o novo arquivo de estilos
import styles from './RifaPage.module.css';

function RifaPage() {
  const { id } = useParams();
  const [rifa, setRifa] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedNumbers, setSelectedNumbers] = useState([]);

  // Estados do formulário
  const [formData, setFormData] = useState({
    nome_comprador: '',
    email_comprador: '',
    telefone_comprador: '',
    cpf_cnpj: '',
  });
  
  const [isReserving, setIsReserving] = useState(false);
  const [reservationError, setReservationError] = useState(null);
  const [pixData, setPixData] = useState(null);

  // useEffect (lógica de busca) permanece o mesmo
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

  // handleNumberClick permanece o mesmo
  const handleNumberClick = (number) => {
    if (selectedNumbers.includes(number)) {
      setSelectedNumbers(selectedNumbers.filter((n) => n !== number));
    } else {
      setSelectedNumbers([...selectedNumbers, number]);
    }
  };

  // handleFormChange permanece o mesmo
  const handleFormChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));
  };

  // handleReservation permanece o mesmo
  const handleReservation = async (e) => {
    e.preventDefault();
    setIsReserving(true);
    setReservationError(null);

    const payload = {
      ...formData,
      numeros: selectedNumbers,
    };

    try {
      const response = await api.post(`/rifas/${id}/reservar`, payload);
      setPixData(response.data);
    } catch (err) {
      let errorMsg = 'Erro ao processar sua reserva. Tente novamente.';
      if (err.response && err.response.data && err.response.data.error) {
        errorMsg = err.response.data.error;
      }
      setReservationError(errorMsg);
      console.error(err);
    } finally {
      setIsReserving(false);
    }
  };

  // 2. Nova função para determinar a CLASSE do botão
  const getButtonClassName = (number) => {
    if (number.status !== 'disponivel') {
      return styles.indisponivel;
    }
    if (selectedNumbers.includes(number.numero)) {
      return styles.selecionado;
    }
    return styles.disponivel;
  };

  // --- RENDERIZAÇÃO COM ESTILOS ---

  if (loading) {
    return <div className="text-center p-10">Carregando detalhes da rifa...</div>;
  }

  if (error) {
    return <div className="text-center p-10 text-red-500">Erro: {error}</div>;
  }

  if (!rifa) {
    return <div>Rifa não encontrada.</div>;
  }

  // 3. Renderização da Tela de PIX (com estilos)
  if (pixData) {
    return (
      <div className={styles.pixContainer}>
        <h2 className={styles.pixTitle}>Pague seu PIX para garantir seus números!</h2>
        <p>Sua reserva foi criada. Realize o pagamento para confirmar.</p>
        <p className={styles.summaryTotal}>
          <strong>Valor Total:</strong> R$ {pixData.value.toFixed(2)}
        </p>
        
        <h3>PIX Copia e Cola:</h3>
        <textarea
          readOnly
          value={pixData.pixQrCode.payload}
          className={styles.pixTextarea}
        />
        <p className={styles.pixStatus}>
          <strong>Status:</strong> {pixData.status}
        </p>
        <a href={pixData.invoiceUrl} target="_blank" rel="noopener noreferrer" className={styles.pixLink}>
          Ver Fatura no Asaas
        </a>
      </div>
    );
  }

  // 4. Renderização da Página Principal (com estilos)
  return (
    <div className={styles.container}>
      <div className={styles.rifaHeader}>
        <h1 className={styles.rifaTitle}>{rifa.titulo}</h1>
        <p className={styles.rifaPrize}><strong>Prêmio:</strong> {rifa.premio}</p>
        <p className={styles.rifaPrice}>R$ {rifa.preco_por_numero.toFixed(2)}</p>
        <p className={styles.rifaDescription}>{rifa.descricao}</p>
      </div>
      
      {/* Carrinho e Formulário */}
      {selectedNumbers.length > 0 && (
        <div className={styles.reservationBox}>
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
          
          <form onSubmit={handleReservation} style={{marginTop: '20px'}}>
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
            
            <button 
              type="submit" 
              disabled={isReserving} 
              className={styles.submitButton}
            >
              {isReserving ? 'Reservando...' : 'Reservar e Gerar PIX'}
            </button>
            
            {reservationError && (
              <p className={styles.errorMessage}>
                <strong>Erro:</strong> {reservationError}
              </p>
            )}
          </form>
        </div>
      )}
      
      {/* Grade de Números */}
      <h2 className={styles.numbersTitle}>Escolha seus números:</h2>
      <div className={styles.numbersGrid}>
        {rifa.numeros.map((numero) => (
          <button 
            key={numero.id} 
            className={`${styles.numberButton} ${getButtonClassName(numero)}`}
            disabled={numero.status !== 'disponivel'}
            onClick={() => handleNumberClick(numero.numero)}
          >
            {numero.numero}
          </button>
        ))}
      </div>
    </div>
  );
}

export default RifaPage;