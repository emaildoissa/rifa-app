// src/pages/AdminFormPage.jsx

import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import api from '../services/api';
import styles from './AdminFormPage.module.css';

function AdminFormPage() {
  const { id } = useParams(); // Pega o ID da URL (se existir)
  const navigate = useNavigate(); // Hook para redirecionar o usuário
  
  const isEditing = Boolean(id); // Se tem ID, estamos editando

  const [formData, setFormData] = useState({
    titulo: '',
    descricao: '',
    premio: '',
    preco_por_numero: 0,
    total_numeros: 100,
    data_sorteio: '',
    status: 'ativa',
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  // Se estamos em modo de edição, busca os dados da rifa ao carregar
  useEffect(() => {
    if (isEditing) {
      setLoading(true);
      api.get(`/rifas/${id}`)
        .then(response => {
          const rifa = response.data;
          // Formata a data para o input type="datetime-local"
          const dataFormatada = rifa.data_sorteio ? new Date(rifa.data_sorteio).toISOString().slice(0, 16) : '';
          
          setFormData({
            titulo: rifa.titulo,
            descricao: rifa.descricao,
            premio: rifa.premio,
            preco_por_numero: rifa.preco_por_numero,
            total_numeros: rifa.total_numeros,
            data_sorteio: dataFormatada,
            status: rifa.status,
          });
        })
        .catch(err => setError('Falha ao carregar dados da rifa.'))
        .finally(() => setLoading(false));
    }
  }, [id, isEditing]);

  const handleChange = (e) => {
    const { name, value, type } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: type === 'number' ? parseFloat(value) : value,
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    // Converte a data local para o formato ISO (UTC) que o Go espera
    const payload = {
      ...formData,
      data_sorteio: formData.data_sorteio ? new Date(formData.data_sorteio).toISOString() : null,
    };

    try {
      if (isEditing) {
        // --- MODO EDIÇÃO (PUT) ---
        // Prepara o payload para o endpoint de ATUALIZAÇÃO (que não permite mudar preços)
        const updatePayload = {
          titulo: payload.titulo,
          descricao: payload.descricao,
          premio: payload.premio,
          data_sorteio: payload.data_sorteio,
          status: payload.status,
        };
        //await api.put(`/rifas/${id}`, updatePayload);
        await api.put(`/admin/rifas/${id}`, updatePayload);
      } else {
        // --- MODO CRIAÇÃO (POST) ---
        //await api.post('/rifas', payload);
        await api.post('/admin/rifas', payload);
      }
      
      // Sucesso! Redireciona para a lista de admin
      navigate('/admin');

    } catch (err) {
      setError('Falha ao salvar a rifa. Verifique os campos e tente novamente.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  if (loading && isEditing) return <div>Carregando formulário...</div>;

  return (
    <div className={styles.container}>
      <h1 className={styles.title}>
        {isEditing ? 'Editar Rifa' : 'Criar Nova Rifa'}
      </h1>

      <form onSubmit={handleSubmit}>
        <div className={styles.formGroup}>
          <label htmlFor="titulo">Título</label>
          <input
            type="text"
            id="titulo"
            name="titulo"
            value={formData.titulo}
            onChange={handleChange}
            required
            className={styles.formInput}
          />
        </div>
        
        <div className={styles.formGroup}>
          <label htmlFor="premio">Prêmio</label>
          <input
            type="text"
            id="premio"
            name="premio"
            value={formData.premio}
            onChange={handleChange}
            required
            className={styles.formInput}
          />
        </div>
        
        <div className={styles.formGroup}>
          <label htmlFor="descricao">Descrição</label>
          <textarea
            id="descricao"
            name="descricao"
            value={formData.descricao}
            onChange={handleChange}
            className={styles.formTextarea}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="preco_por_numero">Preço por Número (R$)</label>
          <input
            type="number"
            id="preco_por_numero"
            name="preco_por_numero"
            value={formData.preco_por_numero}
            onChange={handleChange}
            required
            step="0.01"
            min="0"
            disabled={isEditing} // <-- Importante!
            className={styles.formInput}
          />
        </div>
        
        <div className={styles.formGroup}>
          <label htmlFor="total_numeros">Quantidade Total de Números</label>
          <input
            type="number"
            id="total_numeros"
            name="total_numeros"
            value={formData.total_numeros}
            onChange={handleChange}
            required
            step="1"
            min="1"
            disabled={isEditing} // <-- Importante!
            className={styles.formInput}
          />
        </div>

        <div className={styles.formGroup}>
          <label htmlFor="data_sorteio">Data do Sorteio</label>
          <input
            type="datetime-local"
            id="data_sorteio"
            name="data_sorteio"
            value={formData.data_sorteio}
            onChange={handleChange}
            className={styles.formInput}
          />
        </div>

        {isEditing && ( // Só mostra o status no modo de edição
          <div className={styles.formGroup}>
            <label htmlFor="status">Status</label>
            <select
              id="status"
              name="status"
              value={formData.status}
              onChange={handleChange}
              className={styles.formSelect}
            >
              <option value="ativa">Ativa</option>
              <option value="sorteada">Sorteada</option>
              <option value="cancelada">Cancelada</option>
            </select>
          </div>
        )}

        <button type="submit" className={styles.submitButton} disabled={loading}>
          {loading ? 'Salvando...' : (isEditing ? 'Salvar Alterações' : 'Criar Rifa')}
        </button>

        {error && <p className={styles.errorMessage}>{error}</p>}
      </form>
    </div>
  );
}

export default AdminFormPage;