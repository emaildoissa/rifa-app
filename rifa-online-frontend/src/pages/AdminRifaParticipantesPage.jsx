// src/pages/AdminRifaParticipantesPage.jsx

import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom'; // Import Link para botão de voltar
import api from '../services/api';
import styles from './AdminRifaParticipantesPage.module.css'; // Novo CSS

function AdminRifaParticipantesPage() {
  const { id } = useParams(); // Pega o ID da rifa da URL
  const [participantes, setParticipantes] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [rifaTitulo, setRifaTitulo] = useState(''); // Estado para guardar o título da rifa

  useEffect(() => {
    const fetchData = async () => {
      setLoading(true);
      setError(null);
      try {
        // Primeiro, busca detalhes da rifa para pegar o título (opcional, mas bom para UI)
        try {
          const rifaResponse = await api.get(`/rifas/${id}`);
          setRifaTitulo(rifaResponse.data.titulo);
        } catch (rifaErr) {
          console.warn("Não foi possível buscar o título da rifa, continuando...", rifaErr);
          setRifaTitulo(`Rifa ID ${id}`); // Fallback
        }

        // Agora busca os participantes
        const response = await api.get(`/admin/rifas/${id}/participantes`);
        setParticipantes(response.data);

      } catch (err) {
        setError('Falha ao carregar participantes.');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [id]); // Re-executa se o ID da rifa mudar

  if (loading) return <div>Carregando participantes...</div>;
  if (error) return <div style={{ color: 'red' }}>Erro: {error}</div>;

  return (
    <div className={styles.container}>
      {/* Botão de voltar */}
      <Link to="/admin" className={styles.backButton}>
        &larr; Voltar para Lista de Rifas
      </Link>

      <h1 className={styles.title}>Participantes Confirmados</h1>
      <h2 className={styles.rifaSubTitle}>Rifa: {rifaTitulo}</h2>

      {participantes.length === 0 ? (
        <p className={styles.empty}>Nenhum participante confirmado (pago) para esta rifa ainda.</p>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th>Nome</th>
              <th>Email</th>
              <th>Telefone</th>
              <th>Números Comprados</th>
            </tr>
          </thead>
          <tbody>
            {participantes.map((p, index) => ( // Usar index como key é ok aqui se a lista não mudar dinamicamente
              <tr key={index}>
                <td>{p.nome_comprador}</td>
                <td>{p.email_comprador}</td>
                <td>{p.telefone_comprador}</td>
                {/* Exibe os números separados por vírgula */}
                <td>{p.numeros.join(', ')}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default AdminRifaParticipantesPage;