// src/services/api.js

import axios from 'axios';

// 1. Cria a instância do Axios
const api = axios.create({
  //baseURL: 'http://localhost:8080/api/v1',
  baseURL: 'http://191.252.223.221:8080/api/v1',
});

// 2. Adiciona um "Interceptor" de Requisição
// Isso é uma função que "intercepta" CADA requisição antes dela ser enviada
api.interceptors.request.use(
  (config) => {
    // 3. Pega o token do localStorage
    const token = localStorage.getItem('authToken');
    
    // 4. Se o token existir, anexa ele no cabeçalho Authorization
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`;
    }
    
    // 5. Retorna a configuração modificada para o Axios continuar
    return config;
  },
  (error) => {
    // Em caso de erro ao configurar a requisição
    return Promise.reject(error);
  }
);

export default api;