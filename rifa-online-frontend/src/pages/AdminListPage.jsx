import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import api from '../services/api';
import styles from './AdminListPage.module.css'; // Importe os estilos

function AdminListPage() {
  const [rifas, setRifas] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    fetchAdminRifas();
  }, []);

  const fetchAdminRifas = async () => {
    setLoading(true);
    try {
      // 1. Chame o novo endpoint de admin
      const response = await api.get('/admin/rifas'); 
      setRifas(response.data);
    } catch (err) {
      setError('Falha ao carregar rifas.');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async (rifaId) => {
    // 2. Confirmação antes de apagar
    if (!window.confirm('Tem certeza que deseja apagar esta rifa? Esta ação não pode ser desfeita.')) {
      return;
    }

    try {
      // 3. Chame o endpoint DELETE
      await api.delete(`/admin/rifas/${rifaId}`);
      // 4. Atualize a lista no frontend, removendo a rifa apagada
      setRifas(rifas.filter((rifa) => rifa.id !== rifaId));
    } catch (err) {
      setError('Falha ao apagar rifa. Tente novamente.');
    }
  };

  const handleSortear = async (rifaId, rifaTitulo) => {
    // Confirmação de segurança
    if (!window.confirm(
      `ATENÇÃO!\n\nVocê está prestes a sortear a rifa: "${rifaTitulo}".\n\nEsta ação é IRREVERSÍVEL e definirá um ganhador imediatamente.\n\nDeseja continuar?`
    )) {
      return;
    }

    try {
      // Chama o novo endpoint do backend
      const response = await api.post(`/admin/rifas/${rifaId}/sortear`);
      
      // O backend retorna os dados do ganhador (WinnerInfo)
      const ganhador = response.data;

      // Mostra um alerta de sucesso com os dados do ganhador
      alert(
        `🎉 SORTEIO REALIZADO! 🎉\n\n` +
        `Rifa: ${rifaTitulo}\n` +
        `Número Vencedor: ${ganhador.numero_sorteado}\n\n` +
        `Ganhador:\n` +
        `Nome: ${ganhador.nome_comprador}\n` +
        `Email: ${ganhador.email_comprador}\n` +
        `Telefone: ${ganhador.telefone_comprador}`
      );

      // Atualiza a lista de rifas (a rifa sorteada agora aparecerá como "sorteada")
      fetchAdminRifas();

    } catch (err) {
      // Mostra o erro retornado pela API (ex: "Nenhum número pago", "Já sorteada")
      const errorMsg = err.response?.data?.error || 'Erro desconhecido ao tentar sortear.';
      alert(`Falha no Sorteio: ${errorMsg}`);
    }
  };

  if (loading) return <div>Carregando painel de admin...</div>;
  if (error) return <div style={{ color: 'red' }}>{error}</div>;

  return (
    <div className={styles.container}>
      <div className={styles.header}>
        <h1 className={styles.title}>Painel de Administração</h1>
        <div> {/* --- 1. Adicione um div para agrupar os botões --- */}
          
          {/* --- 2. Adicione este link para a nova página --- */}
          <Link to="/admin/pagamentos" className={styles.pendingButton}>
            Ver Pagamentos Pendentes
          </Link>
          
          <Link to="/admin/new" className={styles.newButton}>
            Criar Nova Rifa
          </Link>
        </div>
      </div>

      <table className={styles.table}>
        <thead>
          <tr>
            <th>ID</th>
            <th>Título</th>
            <th>Status</th>
            <th>Preço</th>
            <th>Progresso</th>
            <th>Ações</th>
          </tr>
        </thead>
        <tbody>
          {rifas.map((rifa) => (
            <tr key={rifa.id}>
              <td>{rifa.id}</td>
              <td>{rifa.titulo}</td>
              <td>{rifa.status}</td>
              <td>R$ {rifa.preco_por_numero.toFixed(2)}</td>
              <td>
                <div className={styles.progressBarContainer}>
                  <div
                    className={styles.progressBarFill}
                    // Calcula a porcentagem e define a largura da barra
                    style={{ width: `${(rifa.numeros_vendidos / rifa.total_numeros) * 100}%` }}
                  ></div>
                </div>
                {/* Texto: "Vendidos / Total" */}
                <span className={styles.progressText}>
                  {rifa.numeros_vendidos} / {rifa.total_numeros}
                </span>
              </td>
              <td className={styles.actions}>
                <Link to={`/admin/edit/${rifa.id}`} className={styles.editButton}>
                  Editar
                </Link>
                <Link to={`/admin/rifa/${rifa.id}/participantes`} className={styles.participantsButton}>
                  Participantes
                </Link>
                {rifa.status === 'ativa' && (
                  <button
                    onClick={() => handleSortear(rifa.id, rifa.titulo)}
                    className={styles.sortearButton}
                  >
                    Sortear
                  </button>
                )}
                <button 
                  onClick={() => handleDelete(rifa.id)} 
                  className={styles.deleteButton}
                >
                  Apagar
                </button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

export default AdminListPage;