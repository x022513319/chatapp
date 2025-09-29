import { useEffect, useState, useRef, useCallback } from "react";
import { useParams, useNavigate } from "react-router-dom";

const BASE_API = import.meta.env.VITE_API_BASE_URL || "http://localhost:8080/api";
const BASE_WS = import.meta.env.VITE_WS_BASE_URL || "ws://localhost:8080";

export default function ChatRoom() {
    const { id } = useParams();
    const navigate = useNavigate();
    const roomId = Number(id) || 1;

    const [rooms, setRooms] = useState([]);
    const [messages, setMessages] = useState([]);
    const [pageInfo, setPageInfo] = useState({ next_before_ts: null, next_before_id: null, has_more: false});
    const [input, setInput] = useState("");
    const [connected, setConnected] = useState(false);

    const listRef = useRef(null);
    const wsRef = useRef(null);
    const bottomRef = useRef(null);
    const token = localStorage.getItem("chatapp_token");

    const autoScrollRef = useRef(true);
    const onScroll = useCallback(() => {
        const el = listRef.current;
        if(!el) return;
        const nearBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80;
        autoScrollRef.current = nearBottom;
    }, []);

    // load room 
    useEffect(() => {
        fetch(`${BASE_API}/rooms`)
            .then(r => r.json())
            .then(setRooms);
    }, []);

    // load latest messages
    const loadLatest = useCallback(async () => {
        const r = await fetch(`${BASE_API}/rooms/${roomId}/messages?limit=50`);
        const data = await r.json();
        const itemAsc = (data?.items ?? []).slice().reverse();
        setMessages(itemAsc);
        setPageInfo({
            next_before_ts: data?.page_info?.next_before_ts ?? null,
            next_before_id: data?.page_info?.next_before_id ?? null,
            has_more: !!data?.page_info?.has_more
        });

        requestAnimationFrame(() => bottomRef.current?.scrollIntoView({ behavior: "auto" }));
    }, [roomId]);

    useEffect(() => {
        loadLatest();
    }, [loadLatest]);

    const loadMore = useCallback(async () => {
        if (!pageInfo.next_before_ts || !pageInfo.next_before_id) return;
        const el = listRef.current;
        const prevHeight = el?.scrollHeight ?? 0;

        const qs = new URLSearchParams({
            before_ts: pageInfo.next_before_ts,
            before_id: String(pageInfo.next_before_id),
            limit: "50"
        });
        const r = await fetch(`${BASE_API}/rooms/${roomId}/messages?${qs.toString()}`);
        const data = await r.json();

        const olderAsc = (data?.items ?? []).slice().reverse();
        setMessages(curr => [...olderAsc, ...curr]);
        setPageInfo({
            next_before_ts: data?.page_info?.next_before_ts ?? null,
            next_before_id: data?.page_info?.next_before_id ?? null,
            has_more: !!data?.page_info?.has_more
        });

        requestAnimationFrame(() => {
            const newHeight = el?.scrollHeight ?? 0;
            if (el) el.scrollTop = newHeight - prevHeight + el.scrollTop;
        });
    }, [pageInfo, roomId]);

    // load ws
    useEffect(() => {
        console.log("WS token =", token);
        if (!token) return;
        let retry = 0;
        let stopped = false;

        function connect() {
            const url = `${BASE_WS}/ws?room_id=${roomId}&token=${encodeURIComponent(token)}`;
            console.log("WS url =", url);
            const ws = new WebSocket(url);
            wsRef.current = ws;

            ws.onopen = () => { setConnected(true); retry = 0; };
            ws.onclose = () => {
                setConnected(false);
                if (stopped) return;
                const delay = Math.min(30000, 1000 * Math.pow(2, retry++));
                setTimeout(connect, delay);
            };
            ws.onerror = (e) => console.log("ws error", e);

            ws.onmessage = (ev) => {
                try {
                    const pkt = JSON.parse(ev.data);
                    if(pkt.type === "message.create") {
                        setMessages(m => [...m, {
                            id: pkt.data.id,
                            user_id: pkt.data.user_id,
                            content: pkt.data.content,
                            created_at: pkt.data.created_at
                        }]);

                        if (autoScrollRef.current) {
                            requestAnimationFrame(() => bottomRef.current?.scrollIntoView({ behavior: "smooth" }));
                        }
                    }
                } catch {}
            };
        }
        
        connect();
        return () => { stopped = true; wsRef.current?.close(); };
    }, [roomId, token]);
        

    useEffect(() => {
        if (autoScrollRef.current) {
            bottomRef.current?.scrollIntoView({ behavior: "smooth" });
        }
    }, [messages]);

    const send = useCallback(() => {
        const v = input.trim();
        if (!v || !wsRef.current) return;
        wsRef.current.send(JSON.stringify({
            type: "message.create",
            data: { content: v }    
        }));
        setInput("");
    }, [input]);

    const onRoomChange = (e) => {
        const nextId = Number(e.target.value);
        navigate(`/rooms/${nextId}`);
    };

    return (
        <div style={{ maxWidth: 720, margin: "32px auto", fontFamily: "system-ui" }}>
        <h1>#chat</h1>

        <div style={{ display: "flex", gap: 8, alignItems: "center", marginBottom: 12 }}>
            <select value={roomId} onChange={onRoomChange}>
            {rooms.map(r => <option key={r.id} value={r.id}>{r.name}</option>)}
            </select>

            {pageInfo.has_more && (
            <button onClick={loadMore} className="border px-2 py-1">載入更舊</button>
            )}

            <span style={{ marginLeft: "auto", fontSize: 12, color: connected ? "green" : "crimson" }}>
            {connected ? "● connected" : "● disconnected"}
            </span>
        </div>

        <div
            ref={listRef}
            onScroll={onScroll}
            style={{ border: "1px solid #ddd", height: 480, overflow: "auto", padding: 12 }}
        >
            {messages.map(m => (
            <div key={m.id} style={{ margin: "6px 0" }}>
                <div style={{ fontSize: 12, color: "#888" }}>
                user {m.user_id} ・ {new Date(m.created_at).toLocaleString()}
                </div>
                <div>{m.content}</div>
            </div>
            ))}
            <div ref={bottomRef} />
        </div>

        <div style={{ display: "flex", gap: 8, marginTop: 12 }}>
            <textarea
            rows={2}
            style={{ flex: 1 }}
            value={input}
            placeholder={connected ? "輸入訊息，Enter（送出Shift+Enter 換行）" : "連線中…"}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => { if (e.key === "Enter" && !e.shiftKey) { e.preventDefault(); send(); } }}
            disabled={!connected}
            />
            <button onClick={send} disabled={!connected || !input.trim()}>Send</button>
        </div>
        </div>
    );
}