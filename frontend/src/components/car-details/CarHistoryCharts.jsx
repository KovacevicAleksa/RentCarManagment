import { useState, useEffect } from "react";
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  Legend,
  ResponsiveContainer,
} from "recharts";
import { Activity, BarChart3, Flame, AlertTriangle, ShieldCheck } from "lucide-react";
import { toChartRows, faultTotal } from "../../lib/events";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

const OVERHEAT_COLOR = "#f97316"; // orange
const CHECK_ENGINE_COLOR = "#f59e0b"; // amber

export default function CarHistoryCharts({ carId }) {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const response = await fetch(
          `${API_URL}/carstats/car/${carId}/events`,
          { credentials: "include" },
        );
        if (!response.ok) throw new Error("Neuspešno učitavanje kvarova");
        const data = await response.json();
        setStats(data);
        setError(null);
        setLoading(false);
      } catch (err) {
        setError(err.message);
        setLoading(false);
      }
    };

    fetchStats();
    const interval = setInterval(fetchStats, 30000);
    return () => clearInterval(interval);
  }, [carId]);

  if (loading) {
    return (
      <div className="mt-8 bg-white rounded-3xl shadow-xl p-12 text-center">
        <Activity className="w-16 h-16 text-gray-400 mx-auto mb-4 animate-pulse" />
        <h3 className="text-xl font-bold text-gray-800 mb-2">
          Učitavanje istorije kvarova...
        </h3>
        <p className="text-gray-600">Preuzimanje podataka o pregrevanju i check engine</p>
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

  const rows = toChartRows(stats?.buckets);
  const overheatTotal = stats?.overheat_total ?? 0;
  const checkEngineTotal = stats?.check_engine_total ?? 0;
  const windowDays = stats?.window_days ?? 30;
  const hasFaults = faultTotal(stats) > 0;

  return (
    <div className="mt-8 space-y-6">
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="bg-gradient-to-br from-orange-500 to-orange-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <Flame className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">{overheatTotal}</span>
          </div>
          <p className="text-sm opacity-90">Pregrevanja</p>
          <p className="text-xs opacity-75 mt-1">u poslednjih {windowDays} dana</p>
        </div>

        <div className="bg-gradient-to-br from-amber-500 to-amber-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <AlertTriangle className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">{checkEngineTotal}</span>
          </div>
          <p className="text-sm opacity-90">Check Engine</p>
          <p className="text-xs opacity-75 mt-1">u poslednjih {windowDays} dana</p>
        </div>

        <div className="bg-gradient-to-br from-red-500 to-red-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <BarChart3 className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">{faultTotal(stats)}</span>
          </div>
          <p className="text-sm opacity-90">Ukupno kvarova</p>
          <p className="text-xs opacity-75 mt-1">pregrevanje + check engine</p>
        </div>

        <div className="bg-gradient-to-br from-emerald-500 to-emerald-600 rounded-2xl shadow-lg p-6 text-white">
          <div className="flex items-center justify-between mb-2">
            <ShieldCheck className="w-8 h-8 opacity-80" />
            <span className="text-2xl font-bold">{rows.length}</span>
          </div>
          <p className="text-sm opacity-90">Dana sa kvarom</p>
          <p className="text-xs opacity-75 mt-1">od {windowDays} praćenih</p>
        </div>
      </div>

      <div className="bg-white rounded-3xl shadow-xl p-6">
        <div className="flex items-center gap-3 mb-6">
          <BarChart3 className="w-6 h-6 text-orange-600" />
          <h3 className="text-xl font-bold text-gray-800">
            Kvarovi tokom vremena (poslednjih {windowDays} dana)
          </h3>
        </div>

        {hasFaults ? (
          <ResponsiveContainer width="100%" height={400}>
            <BarChart data={rows} barGap={4}>
              <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
              <XAxis
                dataKey="label"
                stroke="#6b7280"
                tick={{ fontSize: 12 }}
                interval="preserveStartEnd"
              />
              <YAxis
                stroke="#6b7280"
                tick={{ fontSize: 12 }}
                allowDecimals={false}
                label={{
                  value: "Broj epizoda",
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
              <Bar
                dataKey="overheat"
                fill={OVERHEAT_COLOR}
                name="Pregrevanje"
                radius={[4, 4, 0, 0]}
                maxBarSize={80}
                isAnimationActive={false}
              />
              <Bar
                dataKey="checkEngine"
                fill={CHECK_ENGINE_COLOR}
                name="Check Engine"
                radius={[4, 4, 0, 0]}
                maxBarSize={80}
                isAnimationActive={false}
              />
            </BarChart>
          </ResponsiveContainer>
        ) : (
          <div className="py-16 text-center">
            <ShieldCheck className="w-16 h-16 text-emerald-400 mx-auto mb-4" />
            <h4 className="text-lg font-bold text-gray-800 mb-1">
              Nema zabeleženih kvarova
            </h4>
            <p className="text-gray-600">
              Nijedno pregrevanje ni check engine u poslednjih {windowDays} dana.
            </p>
          </div>
        )}
      </div>
    </div>
  );
}
