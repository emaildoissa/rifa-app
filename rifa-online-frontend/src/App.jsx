// src/App.jsx

import React from 'react';
import { BrowserRouter as Router, Route, Routes } from 'react-router-dom';

// Componentes de Página
import HomePage from './pages/HomePage';
import RifaPage from './pages/RifaPage';
import LoginPage from './pages/LoginPage';
import AdminListPage from './pages/AdminListPage';
import AdminFormPage from './pages/AdminFormPage';
import AdminPagamentosPage from './pages/AdminPagamentosPage';
import AdminRifaParticipantesPage from './pages/AdminRifaParticipantesPage';

// 1. IMPORTE O NOSSO NOVO COMPONENTE DE SEGURANÇA
import ProtectedRoute from './components/ProtectedRoute';

function App() {
  return (
    <div> 
      <Router>
        <Routes>
          {/* --- Rotas Públicas --- */}
          <Route path="/" element={<HomePage />} />
          <Route path="/rifa/:id" element={<RifaPage />} />
          <Route path="/login" element={<LoginPage />} />
          
          {/* --- Rotas de Admin (Agora Protegidas) --- */}
          {/* Agora, o 'element' de cada rota de admin é o nosso 'ProtectedRoute'.
            Passamos a página de admin real (ex: <AdminListPage />) como 'children'.
            O ProtectedRoute vai decidir se renderiza a página ou se redireciona.
          */}
          <Route 
            path="/admin" 
            element={
              <ProtectedRoute>
                <AdminListPage />
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/admin/new" 
            element={
              <ProtectedRoute>
                <AdminFormPage />
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/admin/edit/:id" 
            element={
              <ProtectedRoute>
                <AdminFormPage />
              </ProtectedRoute>
            } 
          />
          <Route 
            path="/admin/pagamentos" 
            element={
              <ProtectedRoute>
                <AdminPagamentosPage />
              </ProtectedRoute>
            } 
          />
          <Route
            path="/admin/rifa/:id/participantes"
            element={
              <ProtectedRoute>
                <AdminRifaParticipantesPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </Router>
    </div>
  );
}

export default App;