import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'
import Dashboard from './pages/Dashboard';
import Register from "./pages/Register";
import Login from './pages/Login';

function App() {
  return (
    <BrowserRouter>
      <nav className="p-3 border-b flex gap-4">
        <Link to="/" className="font-semibold">ChatApp</Link>
        <Link to="/login">Login</Link>
        <Link to="/register">Register</Link>
      </nav>
      <Routes>
        <Route path="/" element={<Dashboard/>} />
        <Route path="/login" element={<Login/>} />
        <Route path="/register" element={<Register/>} />
      </Routes>
    </BrowserRouter>
  );
}

export default App
