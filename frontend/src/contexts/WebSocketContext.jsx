import { createContext, useContext, useEffect, useState, useRef } from "react";
import { fetchUnreadCount, markNotificationsRead } from "../lib/notifications";

const WebSocketContext = createContext(null);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within WebSocketProvider");
  }
  return context;
};

export function WebSocketProvider({ children, apiUrl }) {
  const [cars, setCars] = useState({});
  const [wsConnected, setWsConnected] = useState(false);
  const [unreadCount, setUnreadCount] = useState(0);
  // ringKey increments on every incoming notification; the bell animates when
  // it changes.
  const [ringKey, setRingKey] = useState(0);
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const shouldReconnectRef = useRef(true);

  const WS_URL = apiUrl.replace("http", "ws");

  const handleTelemetry = (payload) => {
    if (!payload || !payload.car_id) return;
    setCars((prev) => ({ ...prev, [payload.car_id]: payload }));
  };

  const handleNotification = (payload) => {
    if (!payload) return;
    setUnreadCount((c) => c + 1);
    setRingKey((k) => k + 1);
  };

  const handleMessage = (raw) => {
    let data;
    try {
      data = JSON.parse(raw);
    } catch {
      console.warn("⚠️ Skipping invalid WS chunk:", raw);
      return;
    }

    if (data.type === "notification") {
      handleNotification(data.payload);
    } else if (data.type === "telemetry") {
      handleTelemetry(data.payload);
    } else if (data.car_id) {
      // Tolerate any non-enveloped telemetry frame.
      handleTelemetry(data);
    }
  };

  const connectWebSocket = () => {
    if (!shouldReconnectRef.current) return;

    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    try {
      const ws = new WebSocket(`${WS_URL}/ws`);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log("✅ WebSocket Connected");
        setWsConnected(true);
        if (reconnectTimeoutRef.current) {
          clearTimeout(reconnectTimeoutRef.current);
          reconnectTimeoutRef.current = null;
        }
      };

      ws.onmessage = (event) => {
        const messages = event.data.split("\n").filter(Boolean);
        for (const msg of messages) {
          handleMessage(msg);
        }
      };

      ws.onerror = (error) => {
        console.error("❌ WebSocket error:", error);
        setWsConnected(false);
      };

      ws.onclose = () => {
        console.log("🔌 WebSocket disconnected");
        setWsConnected(false);
        wsRef.current = null;

        if (shouldReconnectRef.current) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log("🔄 Attempting to reconnect...");
            connectWebSocket();
          }, 3000);
        }
      };
    } catch (error) {
      console.error("Failed to create WebSocket:", error);
      setWsConnected(false);

      if (shouldReconnectRef.current) {
        reconnectTimeoutRef.current = setTimeout(connectWebSocket, 3000);
      }
    }
  };

  useEffect(() => {
    shouldReconnectRef.current = true;
    connectWebSocket();

    return () => {
      shouldReconnectRef.current = false;
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [WS_URL]);

  // Initialise the unread badge from the server, since the WebSocket only
  // delivers notifications that arrive while connected.
  useEffect(() => {
    fetchUnreadCount()
      .then(setUnreadCount)
      .catch(() => {});
  }, []);

  const markAllRead = async () => {
    try {
      await markNotificationsRead();
    } catch {
      // ignore; the badge still resets so the UI stays consistent
    }
    setUnreadCount(0);
  };

  const value = {
    cars,
    wsConnected,
    getCar: (carId) => cars[carId] || null,
    unreadCount,
    ringKey,
    markAllRead,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}
