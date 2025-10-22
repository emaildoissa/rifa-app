// src/components/ProtectedRoute.jsx

import React from 'react';
import { Navigate } from 'react-router-dom';

function ProtectedRoute({ children }) {
  // 1. Verifica se o token existe no localStorage
  const token = localStorage.getItem('authToken');

  // 2. Se não houver token, redireciona para a página de login
  if (!token) {
    // O componente <Navigate> do react-router-dom faz o redirecionamento
    return <Navigate to="/login" replace />;
  }

  // 3. Se houver um token, renderiza o componente filho (a página de admin)
  return children;
}

export default ProtectedRoute;