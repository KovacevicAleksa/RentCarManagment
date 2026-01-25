import { createContext, useContext, useEffect, useState, useRef } from "react";

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
  const wsRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);
  const shouldReconnectRef = useRef(true);

  const WS_URL = apiUrl.replace("http", "ws");

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
          try {
            const telemetry = JSON.parse(msg);
            setCars((prev) => ({
              ...prev,
              [telemetry.car_id]: telemetry,
            }));
          } catch (err) {
            console.warn("⚠️ Skipping invalid WS chunk:", msg);
          }
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

  const value = {
    cars,
    wsConnected,
    getCar: (carId) => cars[carId] || null,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
}
