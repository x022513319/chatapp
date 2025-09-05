import { useState, useEffect } from 'react'

function App() {
  const [msg, setMsg] = useState("...")

  useEffect(() => {
    fetch("http://localhost:8080/health")
      .then(res => res.text())
      .then(setMsg)
  }, [])

  return <h1>Backend says: {msg}</h1>
}

export default App
