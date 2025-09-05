import { useState, useEffect } from 'react'

function App() {
  const [msg, setMsg] = useState("...")

  useEffect(() => {
    fetch("http://localhost:8080/health")
      .then(res => res.text())
      .then(setMsg)
  }, [])

  return (
    <>
      <h1>TEST Backend says: {msg}</h1>
      <p style={{ color: '#16a34a', marginTop: 8 }}>Codex 已成功更新前端 ✅</p>
    </>
  )
}

export default App
