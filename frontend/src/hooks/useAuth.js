import { useEffect, useState } from "react";

const KEY = "chatapp_token";

export function useAuth() {
    const [token, setToken] = useState(localStorage.getItem(KEY) || "");
    useEffect(() => {
        if(token) localStorage.setItem(KEY, token);
        else localStorage.removeItem(KEY);
    }, [token]);
    const logout = () => setToken("");
    return { token, setToken, logout, isAuthed: !!token };
}