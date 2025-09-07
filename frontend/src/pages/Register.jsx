import { useState } from "react";
import { post } from "../api";

export default function Register() {
    const [form, setForm] = useState({ username:"", nickname:"", email:"", password:""})
    const [msg, setMsg] = useState("");

    async function onSubmit(e){
        e.preventDefault();
        const data = await post("/register", form);
        setMsg(JSON.stringify(data));
    }


    return (
        <div className="p-6 max-w-md mx-auto space-y-2">
            <h1 className="text-xl font-bold">Register</h1>
            <form onSubmit={onSubmit} className="space-y-2 border p-4 rounded">
                {["username", "nickname", "email", "password"].map(k => (
                    <input key={k}
                        className="border p-2 w-full"
                        type={k==="password"?"password":"text"}
                        placeholder={k}
                        value={form[k]}
                        onChange={e=>setForm({...form, [k]:e.target.value})}
                    />
                ))}
                <button className="bg-blue-500 text-white px-4 py-2 rounded">Submit</button>
            </form>
            <pre className="text-sm text-black bg-gray-100 p-2">{msg}</pre>
        </div>
    );
}