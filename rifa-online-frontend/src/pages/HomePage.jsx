// src/pages/HomePage.jsx

import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';

// 1. Importe o arquivo de estilos
import styles from './HomePage.module.css';

function HomePage() {
  const [rifas, setRifas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    // ... (lógica do useEffect continua a mesma)
    const fetchRifas = async () => {
      try {
        const response = await api.get('/rifas');
        setRifas(response.data);
      } catch (err) {
        setError('Falha ao carregar as rifas. Tente novamente mais tarde.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchRifas();
  }, []);

  if (loading) {
    return <div className={styles.loading}>Carregando rifas...</div>;
  }

  if (error) {
    return <div className={styles.error}>Erro: {error}</div>;
  }

  return (
    // 2. Aplique as classes usando a sintaxe `styles.nomeDaClasse`
    <div className={styles.container}>
      <h1 className={styles.title}>Rifas Disponíveis</h1>
      
      {rifas.length === 0 ? (
        <p className={styles.empty}>Nenhuma rifa disponível no momento.</p>
      ) : (
        <div className={styles.rifasList}>
          {rifas.map((rifa) => (
            <div key={rifa.id} className={styles.rifaCard}>
              <div className={styles.cardImagePlaceholder}>
                Imagem do Prêmio
              </div>
              
              <div className={styles.cardContent}>
                <h2 className={styles.cardTitle}>{rifa.titulo}</h2>
                <p className={styles.cardPrize}>Prêmio: {rifa.premio}</p>
                <p className={styles.cardPrice}>
                  R$ {rifa.preco_por_numero.toFixed(2)}
                </p>
              </div>
              
              <Link to={`/rifa/${rifa.id}`} className={styles.cardLink}>
                Ver Detalhes e Comprar
              </Link>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

export default HomePage;