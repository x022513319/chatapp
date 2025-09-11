import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Link, Navigate } from 'react-router-dom'
import Dashboard from './pages/Dashboard';
import Register from "./pages/Register";
import Login from './pages/Login';
import ChatRoom from './pages/ChatRoom';

function App() {
  return (
    <BrowserRouter>
      <nav className="p-3 border-b flex gap-4">
        <Link to="/" className="font-semibold">ChatApp</Link>
        <Link to="/login">Login</Link>
        <Link to="/register">Register</Link>
        <Link to="/rooms/1">Chat!</Link>
      </nav>
      <Routes>
        <Route path="/" element={<Dashboard/>} />
        <Route path="/login" element={<Login/>} />
        <Route path="/register" element={<Register/>} />
        <Route path="/rooms/:id" element={ <ChatRoom/> } />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

export default App
