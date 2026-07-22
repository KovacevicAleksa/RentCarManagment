import { useEffect, useState } from "react";
import { ArrowLeft, Bell, AlertTriangle, Info, Car } from "lucide-react";
import { useWebSocket } from "../../contexts/WebSocketContext";
import { fetchNotifications, notificationStyle } from "../../lib/notifications";

const formatTime = (ts) => {
  try {
    return new Date(ts).toLocaleString("sr-RS");
  } catch {
    return "";
  }
};

export default function NotificationsPage({ onBack }) {
  const { markAllRead } = useWebSocket();
  const [items, setItems] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    fetchNotifications()
      .then((data) => {
        if (active) setItems(data);
      })
      .catch((err) => {
        if (active) setError(err.message);
      })
      .finally(() => {
        if (active) setLoading(false);
      });
    // Opening the page marks everything as read and clears the badge.
    markAllRead();
    return () => {
      active = false;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <header className="bg-white shadow-md">
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-md">
                <Bell className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-800">Obaveštenja</h1>
            </div>
            <button
              onClick={onBack}
              className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-xl transition-all"
            >
              <ArrowLeft className="w-4 h-4" />
              Nazad
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 bg-red-50 text-red-700 rounded-xl border border-red-200">
            {error}
          </div>
        )}

        {loading ? (
          <p className="text-gray-500">Učitavanje...</p>
        ) : items.length === 0 ? (
          <div className="bg-white rounded-3xl shadow-xl p-12 text-center">
            <Bell className="w-16 h-16 text-gray-300 mx-auto mb-4" />
            <h3 className="text-xl font-bold text-gray-800 mb-2">
              Nema obaveštenja
            </h3>
            <p className="text-gray-600">
              Ovde će se prikazati nova obaveštenja kada stignu.
            </p>
          </div>
        ) : (
          <ul className="space-y-3">
            {items.map((n) => {
              const style = notificationStyle(n.type);
              const Icon = n.type === "alert" ? AlertTriangle : Info;
              return (
                <li
                  key={n.id}
                  className={`bg-white rounded-2xl shadow-sm border p-5 ${style.card}`}
                >
                  <div className="flex items-start gap-4">
                    <div className="flex items-center justify-center w-10 h-10 rounded-xl bg-white shadow-sm shrink-0">
                      <Icon className={`w-5 h-5 ${style.icon}`} />
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <h3 className="font-semibold text-gray-800">
                          {n.title}
                        </h3>
                        <span
                          className={`text-[10px] font-semibold px-2 py-0.5 rounded-full ${style.badge}`}
                        >
                          {style.label}
                        </span>
                        {n.car_id && (
                          <span className="inline-flex items-center gap-1 text-[10px] font-semibold px-2 py-0.5 rounded-full bg-gray-100 text-gray-600">
                            <Car className="w-3 h-3" />
                            {n.car_id}
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 mt-1">{n.message}</p>
                      <p className="text-xs text-gray-400 mt-2">
                        {formatTime(n.created_at)}
                      </p>
                    </div>
                  </div>
                </li>
              );
            })}
          </ul>
        )}
      </main>
    </div>
  );
}
