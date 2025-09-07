import { useEffect, useState } from "react";
import { useAuth } from "../hooks/useAuth";
import { get } from "../api";
import { Link, useNavigate } from "react-router-dom"


export default function Dashboard() {
    const { token, logout, isAuthed } = useAuth();
    const [me, setMe] = useState(null);
    const nav = useNavigate();

    useEffect(() => {
        if(!isAuthed) { nav("/login"); return; }
        get("/me", token).then(setMe);
    }, [isAuthed, token, nav]);

    if (!isAuthed) return null;

    return (
        <div className="p-6 max-w-2xl mx-auto space-y-3">
            <div className="flex items-center justify-between">
                <h1 className="text-xl font-bold">Dashboard</h1>
                <div className="space-x-3">
                    <Link to="/login" onClick={logout} className="text-red-600">Logout</Link>
                </div>
            </div>
            <pre className="bg-gray-100 text-black p-3 text-sm">{JSON.stringify(me, null, 2)}</pre>
        </div>
    );
}