// src/pages/LoginPage.jsx

import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import api from '../services/api';
import styles from './LoginPage.module.css';

function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState(null);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate(); // Para redirecionar após o login

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      // 1. Chamar o endpoint de login
      const response = await api.post('/login', {
        email: email,
        password: password,
      });

      // 2. Pegar o token da resposta
      const { token } = response.data;

      // 3. Armazenar o token no localStorage
      // O localStorage persiste mesmo se fechar o navegador
      localStorage.setItem('authToken', token);
      
      // 4. Configurar o Axios para usar este token em requisições futuras
      // (Faremos isso no api.js, mas é bom setar aqui também para a primeira vez)
      api.defaults.headers.common['Authorization'] = `Bearer ${token}`;

      // 5. Redirecionar para o painel de admin
      navigate('/admin');

    } catch (err) {
      setError('E-mail ou senha inválidos. Tente novamente.');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className={styles.container}>
      <div className={styles.loginBox}>
        <h1 className={styles.title}>Login do Admin</h1>
        <form onSubmit={handleSubmit}>
          <div className={styles.formGroup}>
            <label htmlFor="email">Email</label>
            <input
              type="email"
              id="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              className={styles.formInput}
            />
          </div>
          <div className={styles.formGroup}>
            <label htmlFor="password">Senha</label>
            <input
              type="password"
              id="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              className={styles.formInput}
            />
          </div>
          <button type="submit" className={styles.submitButton} disabled={loading}>
            {loading ? 'Entrando...' : 'Entrar'}
          </button>
          {error && <p className={styles.errorMessage}>{error}</p>}
        </form>
      </div>
    </div>
  );
}

export default LoginPage;