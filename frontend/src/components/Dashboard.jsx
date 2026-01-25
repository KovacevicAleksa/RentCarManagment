import { useState, useEffect } from "react";
import { Car, Gauge, Thermometer, Droplet, MapPin, Map } from "lucide-react";
import { useWebSocket } from "../contexts/WebSocketContext";
import CarDetails from "./car-details/CarDetails";
import FleetMap from "./fleet-map/FleetMap";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

export default function Dashboard({ onLogout }) {
  const { cars, wsConnected } = useWebSocket();
  const [userInfo, setUserInfo] = useState(null);
  const [showUserMenu, setShowUserMenu] = useState(false);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [selectedCar, setSelectedCar] = useState(null);
  const [showFleetMap, setShowFleetMap] = useState(false);

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

  const formatTimestamp = (timestamp) => {
    return new Date(timestamp).toLocaleTimeString("sr-RS");
  };

  // Ako je vozilo selektovano, prikaži CarDetails komponentu
  if (selectedCar) {
    return (
      <CarDetails carId={selectedCar} onBack={() => setSelectedCar(null)} />
    );
  }

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
              <div className="ml-4 flex items-center gap-2">
                <div
                  className={`w-2 h-2 rounded-full ${
                    wsConnected ? "bg-green-500 animate-pulse" : "bg-red-500"
                  }`}
                />
                <span className="text-xs text-gray-600">
                  {wsConnected ? "Live" : "Offline"}
                </span>
              </div>
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
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h2 className="text-2xl font-bold text-gray-800 mb-2">
              Kontrolna tabla voznog parka
            </h2>
            <p className="text-gray-600">
              Praćenje vozila u realnom vremenu ({Object.keys(cars).length}{" "}
              aktivnih)
            </p>
          </div>

          {Object.keys(cars).length > 0 && (
            <button
              onClick={() => setShowFleetMap(true)}
              className="flex items-center gap-2 px-6 py-3 bg-gradient-to-r from-blue-600 to-purple-600 text-white rounded-xl hover:shadow-lg transform hover:scale-105 transition-all duration-200 font-medium"
            >
              <Map className="w-5 h-5" />
              Prikaži sve na mapi
            </button>
          )}
        </div>

        {error && (
          <div className="mb-6 p-4 bg-red-50 text-red-700 rounded-xl border border-red-200">
            {error}
          </div>
        )}

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {Object.values(cars).map((car) => (
            <div
              key={car.car_id}
              onClick={() => setSelectedCar(car.car_id)}
              className="bg-white rounded-2xl shadow-lg p-6 hover:shadow-xl transition-all cursor-pointer transform hover:scale-105 duration-200"
            >
              <div className="flex items-center justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="flex items-center justify-center w-12 h-12 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-md">
                    <Car className="w-6 h-6 text-white" />
                  </div>
                  <div>
                    <h3 className="text-lg font-bold text-gray-800">
                      {car.car_id}
                    </h3>
                    <p className="text-xs text-gray-500">
                      {formatTimestamp(car.timestamp)}
                    </p>
                  </div>
                </div>
              </div>

              <div className="space-y-3">
                <div
                  className={`flex items-center justify-between p-3 rounded-lg ${
                    car.engine_temperature >= 100
                      ? "bg-red-50"
                      : car.engine_temperature >= 90
                        ? "bg-yellow-50"
                        : "bg-green-50"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Thermometer
                      className={`w-4 h-4 ${
                        car.engine_temperature >= 100
                          ? "text-red-600"
                          : car.engine_temperature >= 90
                            ? "text-yellow-600"
                            : "text-green-600"
                      }`}
                    />
                    <span className="text-sm font-medium text-gray-700">
                      Motor
                    </span>
                  </div>
                  <span
                    className={`text-sm font-bold ${
                      car.engine_temperature >= 100
                        ? "text-red-600"
                        : car.engine_temperature >= 90
                          ? "text-yellow-600"
                          : "text-green-600"
                    }`}
                  >
                    {car.engine_temperature.toFixed(1)}°C
                  </span>
                </div>

                <div
                  className={`flex items-center justify-between p-3 rounded-lg ${
                    car.engine_coolant_temp >= 95
                      ? "bg-red-50"
                      : car.engine_coolant_temp >= 85
                        ? "bg-yellow-50"
                        : "bg-green-50"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Droplet
                      className={`w-4 h-4 ${
                        car.engine_coolant_temp >= 95
                          ? "text-red-600"
                          : car.engine_coolant_temp >= 85
                            ? "text-yellow-600"
                            : "text-green-600"
                      }`}
                    />
                    <span className="text-sm font-medium text-gray-700">
                      Rashladna
                    </span>
                  </div>
                  <span
                    className={`text-sm font-bold ${
                      car.engine_coolant_temp >= 95
                        ? "text-red-600"
                        : car.engine_coolant_temp >= 85
                          ? "text-yellow-600"
                          : "text-green-600"
                    }`}
                  >
                    {car.engine_coolant_temp.toFixed(1)}°C
                  </span>
                </div>

                <div className="flex items-center justify-between p-3 bg-blue-50 rounded-lg">
                  <div className="flex items-center gap-2">
                    <Gauge className="w-4 h-4 text-blue-600" />
                    <span className="text-sm font-medium text-gray-700">
                      Gas
                    </span>
                  </div>
                  <span className="text-sm font-bold text-blue-600">
                    {(car.gas_throttle * 100).toFixed(0)}%
                  </span>
                </div>

                <div
                  className={`flex items-center justify-between p-3 rounded-lg ${
                    car.fuel_level <= 15
                      ? "bg-red-50"
                      : car.fuel_level <= 30
                        ? "bg-yellow-50"
                        : "bg-green-50"
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Droplet
                      className={`w-4 h-4 ${
                        car.fuel_level <= 15
                          ? "text-red-600"
                          : car.fuel_level <= 30
                            ? "text-yellow-600"
                            : "text-green-600"
                      }`}
                    />
                    <span className="text-sm font-medium text-gray-700">
                      Gorivo
                    </span>
                  </div>
                  <span
                    className={`text-sm font-bold ${
                      car.fuel_level <= 15
                        ? "text-red-600"
                        : car.fuel_level <= 30
                          ? "text-yellow-600"
                          : "text-green-600"
                    }`}
                  >
                    {car.fuel_level.toFixed(1)}%
                  </span>
                </div>

                <div className="flex items-center justify-between p-3 bg-purple-50 rounded-lg">
                  <div className="flex items-center gap-2">
                    <MapPin className="w-4 h-4 text-purple-600" />
                    <span className="text-sm font-medium text-gray-700">
                      Lokacija
                    </span>
                  </div>
                  <span className="text-xs font-mono text-purple-600">
                    {car.latitude.toFixed(4)}, {car.longitude.toFixed(4)}
                  </span>
                </div>
              </div>

              <div className="mt-4 pt-4 border-t border-gray-200">
                <p className="text-xs text-center text-gray-500">
                  Klikni za detaljne informacije
                </p>
              </div>
            </div>
          ))}
        </div>

        {Object.keys(cars).length === 0 && (
          <div className="bg-white rounded-3xl shadow-xl p-12 text-center">
            <Car className="w-16 h-16 text-gray-400 mx-auto mb-4" />
            <h3 className="text-xl font-bold text-gray-800 mb-2">
              Nema aktivnih vozila
            </h3>
            <p className="text-gray-600">Čekanje telemetrijskih podataka...</p>
          </div>
        )}
      </main>

      {/* Fleet Map Modal */}
      {showFleetMap && (
        <FleetMap cars={cars} onClose={() => setShowFleetMap(false)} />
      )}
    </div>
  );
}
