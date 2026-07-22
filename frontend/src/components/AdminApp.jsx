import { useState, useEffect } from "react";
import { ShieldCheck, ShieldAlert, LogOut } from "lucide-react";
import Login from "./Login";
import AdminPanel from "./admin/AdminPanel";
import { WebSocketProvider } from "../contexts/WebSocketContext";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

export default function AdminApp() {
  const [checkingAuth, setCheckingAuth] = useState(true);
  const [userInfo, setUserInfo] = useState(null);

  const fetchUser = async () => {
    try {
      const res = await fetch(`${API_URL}/auth/me`, { credentials: "include" });
      if (res.ok) {
        setUserInfo(await res.json());
      } else {
        setUserInfo(null);
      }
    } catch {
      setUserInfo(null);
    } finally {
      setCheckingAuth(false);
    }
  };

  useEffect(() => {
    fetchUser();
  }, []);

  const handleLogout = async () => {
    try {
      await fetch(`${API_URL}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // ignore network errors on logout
    } finally {
      setUserInfo(null);
    }
  };

  if (checkingAuth) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-100 via-white to-slate-200 flex items-center justify-center">
        <div className="text-gray-600">Učitavanje...</div>
      </div>
    );
  }

  // Not logged in → admin-branded login
  if (!userInfo) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-100 via-white to-slate-200 flex items-center justify-center p-4">
        <div className="w-full max-w-md">
          <div className="text-center mb-8">
            <div className="inline-flex items-center justify-center w-16 h-16 bg-gradient-to-br from-slate-700 to-slate-900 rounded-2xl mb-4 shadow-lg">
              <ShieldCheck className="w-8 h-8 text-white" />
            </div>
            <h1 className="text-3xl font-bold text-gray-800">RentCar Admin</h1>
            <p className="text-gray-600 mt-2">Prijava za administratore</p>
          </div>
          <div className="bg-white rounded-3xl shadow-xl p-8">
            <Login onLoginSuccess={fetchUser} />
          </div>
        </div>
      </div>
    );
  }

  // Logged in but not an admin → access denied
  if (userInfo.role !== "admin") {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-100 via-white to-slate-200 flex items-center justify-center p-4">
        <div className="bg-white rounded-3xl shadow-xl p-12 text-center max-w-md">
          <div className="inline-flex items-center justify-center w-16 h-16 bg-red-100 rounded-2xl mb-4">
            <ShieldAlert className="w-8 h-8 text-red-600" />
          </div>
          <h1 className="text-2xl font-bold text-gray-800 mb-2">Pristup odbijen</h1>
          <p className="text-gray-600 mb-6">
            Nalog <strong>{userInfo.email}</strong> nema administratorska prava.
          </p>
          <button
            onClick={handleLogout}
            className="inline-flex items-center gap-2 px-6 py-3 bg-red-50 text-red-600 rounded-xl font-semibold hover:bg-red-100 transition-all"
          >
            <LogOut className="w-4 h-4" />
            Odjavi se
          </button>
        </div>
      </div>
    );
  }

  // Admin → the admin panel, with a live WebSocket connection so the admin
  // receives notifications like any other registered user.
  return (
    <WebSocketProvider apiUrl={API_URL}>
      <AdminPanel
        onLogout={handleLogout}
        userEmail={userInfo.email}
        currentUserId={userInfo.user_id}
      />
    </WebSocketProvider>
  );
}
