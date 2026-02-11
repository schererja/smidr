import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import SystemList from './pages/SystemList';
import SystemDetail from './pages/SystemDetail';
import './App.css';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className="min-h-screen flex flex-col bg-gray-50">
        <header className="bg-gradient-to-r from-purple-600 to-purple-800 text-white shadow-md">
          <div className="max-w-7xl mx-auto px-6 py-6">
            <h1 className="text-3xl font-bold tracking-tight">Smidr</h1>
            <p className="text-sm opacity-90 mt-1">System Monitoring & Infrastructure Delivery</p>
          </div>
        </header>
        <main className="flex-1 pt-8">
          <Routes>
            <Route path="/" element={<Navigate to="/systems" replace />} />
            <Route path="/systems" element={<SystemList />} />
            <Route path="/systems/:agentId" element={<SystemDetail />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
};

export default App;
