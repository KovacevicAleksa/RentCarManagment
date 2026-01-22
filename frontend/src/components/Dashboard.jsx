import { useState, useEffect } from "react";
import { Car } from "lucide-react";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

export default function Dashboard({ onLogout }) {
  const [userInfo, setUserInfo] = useState(null);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    fetchUserInfo();
  }, []);

  const fetchUserInfo = async () => {
    try {
      const res = await fetch(`${API_URL}/auth/me`, {
        credentials: "include",
      });

      if (!res.ok) {
        throw new Error("Neuspešno učitavanje informacija");
      }

      const data = await res.json();
      setUserInfo(data);
      setLoading(false);
    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  };

  const handleLogout = async () => {
    try {
      await fetch(`${API_URL}/auth/logout`, {
        method: "POST",
        credentials: "include",
      });
    } catch (err) {
      console.error("Greška pri odjavi:", err);
    } finally {
      onLogout();
    }
  };

  const getInitials = (email) => {
    if (!email) return "U";
    return email.charAt(0).toUpperCase();
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <header className="bg-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-md">
                <Car className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-800">RentCar</h1>
            </div>

            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-3 px-3 py-2 rounded-xl hover:bg-gray-100 transition-all"
              >
                <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-full shadow-md">
                  <span className="text-white font-semibold text-sm">
                    {loading ? "..." : getInitials(userInfo?.email)}
                  </span>
                </div>
                <div className="hidden sm:block text-left">
                  <p className="text-sm font-medium text-gray-700">
                    {loading ? "Učitavanje..." : userInfo?.email || "Korisnik"}
                  </p>
                  <p className="text-xs text-gray-500">
                    ID: {userInfo?.user_id || "-"}
                  </p>
                </div>
              </button>

              {showUserMenu && (
                <div className="absolute right-0 mt-2 w-64 bg-white rounded-xl shadow-xl border border-gray-200 overflow-hidden z-10">
                  <div className="p-4 bg-gradient-to-br from-blue-50 to-purple-50 border-b border-gray-200">
                    <div className="flex items-center gap-3">
                      <div className="flex items-center justify-center w-12 h-12 bg-gradient-to-br from-blue-600 to-purple-600 rounded-full shadow-md">
                        <span className="text-white font-semibold">
                          {getInitials(userInfo?.email)}
                        </span>
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-semibold text-gray-800 truncate">
                          {userInfo?.email || "Korisnik"}
                        </p>
                        <p className="text-xs text-gray-600">
                          Korisnički ID: {userInfo?.user_id}
                        </p>
                      </div>
                    </div>
                  </div>

                  <div className="p-2">
                    <button
                      onClick={handleLogout}
                      className="w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-red-50 rounded-lg transition-all font-medium"
                    >
                      Odjavi se
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white rounded-3xl shadow-xl p-8">
          <h2 className="text-2xl font-bold text-gray-800 mb-4">Dobrodošli!</h2>
          <p className="text-gray-600">
            Uspešno ste se prijavili u RentCar sistem.
          </p>

          {error && (
            <div className="mt-4 p-4 bg-red-50 text-red-700 rounded-xl border border-red-200">
              {error}
            </div>
          )}

          {userInfo && (
            <div className="mt-6 p-6 bg-gradient-to-br from-blue-50 to-purple-50 rounded-2xl border border-gray-200">
              <h3 className="text-lg font-semibold text-gray-800 mb-3">
                Vaše informacije
              </h3>
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">Email:</span>
                  <span className="text-sm font-medium text-gray-800">
                    {userInfo.email}
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm text-gray-600">Korisnički ID:</span>
                  <span className="text-sm font-medium text-gray-800">
                    {userInfo.user_id}
                  </span>
                </div>
              </div>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}
