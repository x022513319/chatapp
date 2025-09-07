import { useState } from "react";
import { post } from "../api";
import { useAuth } from "../hooks/useAuth";
import { useNavigate, Link } from "react-router-dom"

export default function Login() {
    const [email, setEmail] = useState("");
    const [password, setPassword] = useState("");
    const [msg, setMsg] = useState("");
    const { setToken } = useAuth();
    const nav = useNavigate();

    async function onSubmit(e){
        e.preventDefault()
        const data = await post("/login", { email, password });
        if(data.access_token) {
            setToken(data.access_token);
            nav("/");
        } else {
            setMsg(JSON.stringify(data));
        }
    }

    return (
        <div className="p-6 max-w-wd mx-auto space-y-2">
            <h1 className="text-xl font-bold">Login</h1>
            <form onSubmit={onSubmit} className="space-y-2 border p-4 rounded">
                <input className="border p-2 w-full" placeholder="Email"
                    value={email} onChange={e=>setEmail(e.target.value)} />
                <input className="border p-2 w-full" type="password" placeholder="Password" 
                    value={password} onChange={e=>setPassword(e.target.value)} />
                <button className="bg-green-600 text-white px-4 py-2 rounded">Login</button>
            </form>
            <div className="text-sm">No account? <Link className="text-blue-600" to="/register">Register</Link></div>
            <pre className="text-sm text-black bg-gray-100 p-2">{msg}</pre>
        </div>
    );
}