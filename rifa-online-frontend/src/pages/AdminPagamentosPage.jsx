// src/pages/AdminPagamentosPage.jsx

import React, { useState, useEffect } from 'react';
import api from '../services/api';
import styles from './AdminPagamentosPage.module.css'; // Usaremos um novo CSS

function AdminPagamentosPage() {
  const [pagamentos, setPagamentos] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchPagamentos();
  }, []);

  const fetchPagamentos = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get('/admin/pagamentos/pendentes');
      setPagamentos(response.data);
    } catch (err) {
      setError('Falha ao carregar pagamentos pendentes.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleAprovar = async (id) => {
    if (!window.confirm('Tem certeza que deseja APROVAR este pagamento? Esta ação enviará o e-mail de confirmação.')) {
      return;
    }
    try {
      await api.post(`/admin/pagamentos/${id}/aprovar`);
      // Remove o pagamento da lista local
      setPagamentos(pagamentos.filter((p) => p.pagamento_id !== id));
    } catch (err) {
      alert('Erro ao aprovar pagamento: ' + (err.response?.data?.error || 'Erro desconhecido'));
    }
  };

  const handleLiberar = async (id) => {
    if (!window.confirm('Tem certeza que deseja LIBERAR (cancelar) esta reserva? Os números voltarão a ficar disponíveis.')) {
      return;
    }
    try {
      await api.post(`/admin/pagamentos/${id}/liberar`);
      // Remove o pagamento da lista local
      setPagamentos(pagamentos.filter((p) => p.pagamento_id !== id));
    } catch (err) {
      alert('Erro ao liberar números: ' + (err.response?.data?.error || 'Erro desconhecido'));
    }
  };

  // Helper para formatar data
  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleString('pt-BR');
  };

  if (loading) return <div>Carregando pagamentos pendentes...</div>;
  if (error) return <div style={{ color: 'red' }}>{error}</div>;

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>Gerenciador de Pagamentos Pendentes</h1>

      {pagamentos.length === 0 ? (
        <p className={styles.empty}>Nenhum pagamento pendente no momento.</p>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th>ID Pag.</th>
              <th>Comprador</th>
              <th>Email / Telefone</th>
              <th>Rifa</th>
              <th>Números</th>
              <th>Valor (R$)</th>
              <th>Data Reserva</th>
              <th>Ações</th>
            </tr>
          </thead>
          <tbody>
            {pagamentos.map((pg) => (
              <tr key={pg.pagamento_id}>
                <td>{pg.pagamento_id}</td>
                <td>{pg.nome_comprador}</td>
                <td>
                  <div>{pg.email_comprador}</div>
                  <div>{pg.telefone_comprador}</div>
                </td>
                <td>{pg.rifa_titulo}</td>
                <td>{pg.numeros.join(', ')}</td>
                <td>{pg.valor_total.toFixed(2)}</td>
                <td>{formatDate(pg.data_reserva)}</td>
                <td className={styles.actions}>
                  <button
                    onClick={() => handleAprovar(pg.pagamento_id)}
                    className={styles.approveButton}
                  >
                    Aprovar
                  </button>
                  <button
                    onClick={() => handleLiberar(pg.pagamento_id)}
                    className={styles.releaseButton}
                  >
                    Liberar
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default AdminPagamentosPage;