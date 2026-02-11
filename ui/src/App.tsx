import React, { useState } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import SystemList from './pages/SystemList';
import SystemDetail from './pages/SystemDetail';
import Sidebar from './components/Sidebar';
import Header from './components/Header';
import './App.css';

const App: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState(false);

  return (
    <BrowserRouter>
      <div className="min-h-screen flex bg-gray-50">
        <Sidebar isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 flex flex-col w-full lg:w-auto">
          <Header onMenuClick={() => setSidebarOpen(!sidebarOpen)} />
          <main className="flex-1 overflow-auto">
            <Routes>
              <Route path="/" element={<Navigate to="/systems" replace />} />
              <Route path="/systems" element={<SystemList />} />
              <Route path="/systems/:agentId" element={<SystemDetail />} />
              <Route path="/settings" element={<div className="p-8 text-gray-500">Settings coming soon...</div>} />
            </Routes>
          </main>
        </div>
      </div>
    </BrowserRouter>
  );
};

export default App;
