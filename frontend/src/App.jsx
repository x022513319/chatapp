import { useState, useEffect } from 'react'

function App() {
  const [token, setToken] = useState("")
  const [me, setMe] = useState(null)

  const baseUrl = import.meta.env.VITE_API_BASE_URL

  async function register() {
    const res = await fetch(`${baseUrl}/register`, {
      method: "POST",
      headers: { "Content-Type": "application/json"},
      body: JSON.stringify({
        username: "alice",
        nickname: "Alice",
        email: "alice@example.com",
        password: "p@ssw0rd123"
      })
    })
    console.log("register", await res.json())
  }

  async function login() {
    const res = await fetch(`${baseUrl}/login`,{
      method: "POST",
      headers: { "Content-Type": "application/json"},
      body: JSON.stringify({
        email: "alice@example.com",
        password: "p@ssw0rd1234ˋ"
      })
    })
    const data = await res.json()
    console.log("login:", data)
    setToken(data.access_token)
  }

  async function getMe(){
    const res = await fetch(`${baseUrl}/me`, {
      headers: { Authorization: `Bearer ${token}` }
    })
    const data = await res.json()
    console.log("me:", data)
    setMe(data)
  }

  return (
    <div style={{ padding: 20 }}>
      <button onClick={register}>Register</button>
      <button onClick={login}>Login</button>
      <button onClick={getMe} disabled={!token}>Get Me</button>

      {me && (
        <pre>{JSON.stringify(me, null, 2)}</pre>
      )}
    </div>
  )
}

export default App
