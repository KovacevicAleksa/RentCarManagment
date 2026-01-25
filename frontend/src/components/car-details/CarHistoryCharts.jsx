import { useState, useEffect } from "react";
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
  AreaChart,
  Area,
} from "recharts";
import { TrendingDown, MapPin, Activity, BarChart3 } from "lucide-react";

export default function CarHistoryCharts({ carId }) {
  const [historyData, setHistoryData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchHistory = async () => {
      try {
        setLoading(true);
        const response = await fetch(
          `http://localhost:8010/history/car/${carId}?limit=50`,
          {
            credentials: "include",
          },
        );
        if (!response.ok) throw new Error("Failed to fetch history");
        const data = await response.json();
        setHistoryData(data);
        setError(null);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    fetchHistory();
    const interval = setInterval(fetchHistory, 30000);
    return () => clearInterval(interval);
  }, [carId]);

  if (loading) {
    return (
      <div className="mt-8 bg-white rounded-3xl shadow-xl p-12 text-center">
        <Activity className="w-16 h-16 text-gray-400 mx-auto mb-4 animate-pulse" />
        <h3 className="text-xl font-bold text-gray-800 mb-2">
          Učitavanje istorije...
        </h3>
        <p className="text-gray-600">Preuzimanje telemetrijskih podataka</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="mt-8 bg-red-50 rounded-3xl shadow-xl p-12 text-center">
        <Activity className="w-16 h-16 text-red-400 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-red-800 mb-2">Greška</h3>
        <p className="text-red-600">{error}</p>
      </div>
    );
  }

  if (
    !historyData ||
    !historyData.records ||
    historyData.records.length === 0
  ) {
    return (
      <div className="mt-8 bg-white rounded-3xl shadow-xl p-12 text-center">
        <BarChart3 className="w-16 h-16 text-gray-400 mx-auto mb-4" />
        <h3 className="text-xl font-bold text-gray-800 mb-2">Nema podataka</h3>
        <p className="text-gray-600">Nema istorijskih podataka za prikaz</p>
      </div>
    );
  }

  const chartData = [...historyData.records].reverse().map((record, index) => ({
    index: index + 1,
    fuel: parseFloat(record.fuel?.toFixed(2) || 0),
    time: new Date(record.timestamp).toLocaleTimeString("sr-RS", {
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
    }),
    fullTimestamp: record.timestamp,
  }));

  const fuelValues = chartData.map((d) => d.fuel);
  const totalFuelConsumed = fuelValues[0] - fuelValues[fuelValues.length - 1];
  const avgFuelLevel =
    fuelValues.reduce((a, b) => a + b, 0) / fuelValues.length;
  const minFuel = Math.min(...fuelValues);
  const maxFuel = Math.max(...fuelValues);

  const calculateDistance = () => {
    let totalDistance = 0;
    for (let i = 1; i < historyData.records.length; i++) {
      const prev = historyData.records[i - 1];
      const curr = historyData.records[i];
      if (prev.latitude && prev.longitude && curr.latitude && curr.longitude) {
        const lat1 = (prev.latitude * Math.PI) / 180;
        const lat2 = (curr.latitude * Math.PI) / 180;
        const dLat = ((curr.latitude - prev.latitude) * Math.PI) / 180;
        const dLon = ((curr.longitude - prev.longitude) * Math.PI) / 180;

        const a =
          Math.sin(dLat / 2) * Math.sin(dLat / 2) +
          Math.cos(lat1) *
            Math.cos(lat2) *
            Math.sin(dLon / 2) *
            Math.sin(dLon / 2);
        const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
        totalDistance += 6371 * c;
      }
    }
    return totalDistance;
  };

  const estimatedDistance = calculateDistance();

  return (
    <div className="mt-8 space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <TrendingDown className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">
              {totalFuelConsumed.toFixed(2)}%
            </span>
          </div>
          <p className="text-sm opacity-90">Potrošeno goriva</p>
          <p className="text-xs opacity-75 mt-1">u periodu praćenja</p>
        </div>

        <div className="bg-gradient-to-br from-green-500 to-green-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <Activity className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">
              {avgFuelLevel.toFixed(1)}%
            </span>
          </div>
          <p className="text-sm opacity-90">Prosečan nivo</p>
          <p className="text-xs opacity-75 mt-1">goriva u rezervoaru</p>
        </div>

        <div className="bg-gradient-to-br from-purple-500 to-purple-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <MapPin className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">
              {estimatedDistance.toFixed(2)}
            </span>
          </div>
          <p className="text-sm opacity-90">Pređeno km</p>
          <p className="text-xs opacity-75 mt-1">procenjena razdaljina</p>
        </div>

        <div className="bg-gradient-to-br from-orange-500 to-orange-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <BarChart3 className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">{historyData.count}</span>
          </div>
          <p className="text-sm opacity-90">Ukupno zapisa</p>
          <p className="text-xs opacity-75 mt-1">telemetrijskih podataka</p>
        </div>
      </div>

      <div className="bg-white rounded-3xl shadow-xl p-6">
        <div className="flex items-center gap-3 mb-6">
          <TrendingDown className="w-6 h-6 text-blue-600" />
          <h3 className="text-xl font-bold text-gray-800">Potrošnja goriva</h3>
        </div>
        <ResponsiveContainer width="100%" height={300}>
          <AreaChart data={chartData}>
            <defs>
              <linearGradient id="fuelGradient" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.8} />
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.1} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis
              dataKey="time"
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              interval="preserveStartEnd"
            />
            <YAxis
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              domain={[Math.floor(minFuel - 1), Math.ceil(maxFuel + 1)]}
              label={{
                value: "Gorivo (%)",
                angle: -90,
                position: "insideLeft",
                style: { fontSize: 12 },
              }}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "rgba(255, 255, 255, 0.95)",
                border: "1px solid #e5e7eb",
                borderRadius: "12px",
                padding: "12px",
              }}
              labelStyle={{ fontWeight: "bold", marginBottom: "8px" }}
            />
            <Legend wrapperStyle={{ paddingTop: "20px" }} iconType="circle" />
            <Area
              type="monotone"
              dataKey="fuel"
              stroke="#3b82f6"
              strokeWidth={3}
              fill="url(#fuelGradient)"
              name="Nivo goriva (%)"
              animationDuration={1000}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-3xl shadow-xl p-6">
        <div className="flex items-center gap-3 mb-6">
          <Activity className="w-6 h-6 text-purple-600" />
          <h3 className="text-xl font-bold text-gray-800">
            Trend nivoa goriva
          </h3>
        </div>
        <ResponsiveContainer width="100%" height={250}>
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
            <XAxis
              dataKey="index"
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              label={{
                value: "Broj merenja",
                position: "insideBottom",
                offset: -5,
                style: { fontSize: 12 },
              }}
            />
            <YAxis
              stroke="#6b7280"
              tick={{ fontSize: 12 }}
              domain={[Math.floor(minFuel - 1), Math.ceil(maxFuel + 1)]}
              label={{
                value: "Gorivo (%)",
                angle: -90,
                position: "insideLeft",
                style: { fontSize: 12 },
              }}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: "rgba(255, 255, 255, 0.95)",
                border: "1px solid #e5e7eb",
                borderRadius: "12px",
                padding: "12px",
              }}
            />
            <Legend wrapperStyle={{ paddingTop: "20px" }} iconType="line" />
            <Line
              type="monotone"
              dataKey="fuel"
              stroke="#a855f7"
              strokeWidth={3}
              dot={{ fill: "#a855f7", r: 4 }}
              activeDot={{ r: 6 }}
              name="Nivo goriva (%)"
              animationDuration={1000}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white rounded-3xl shadow-xl p-6">
        <div className="flex items-center gap-3 mb-6">
          <BarChart3 className="w-6 h-6 text-green-600" />
          <h3 className="text-xl font-bold text-gray-800">Sumarni podaci</h3>
        </div>
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-gray-50 rounded-xl p-4">
            <p className="text-sm text-gray-600 mb-1">Maksimalni nivo</p>
            <p className="text-2xl font-bold text-gray-800">
              {maxFuel.toFixed(2)}%
            </p>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <p className="text-sm text-gray-600 mb-1">Minimalni nivo</p>
            <p className="text-2xl font-bold text-gray-800">
              {minFuel.toFixed(2)}%
            </p>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <p className="text-sm text-gray-600 mb-1">Raspon</p>
            <p className="text-2xl font-bold text-gray-800">
              {(maxFuel - minFuel).toFixed(2)}%
            </p>
          </div>
          <div className="bg-gray-50 rounded-xl p-4">
            <p className="text-sm text-gray-600 mb-1">Merenja</p>
            <p className="text-2xl font-bold text-gray-800">
              {chartData.length}
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
