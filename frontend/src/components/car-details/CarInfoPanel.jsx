import {
  Clock,
  Activity,
  Thermometer,
  Droplet,
  Gauge,
  AlertTriangle,
  CheckCircle,
  Flame,
  ShieldCheck,
} from "lucide-react";
import {
  reliabilityLevel,
  formatScore,
  checkEngineCountClasses,
} from "../../lib/reliability";
import { LEVEL_STYLES } from "../../lib/status";

const formatTimestamp = (timestamp) => {
  return new Date(timestamp).toLocaleString("sr-RS", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  });
};

const getStatusColor = (value, thresholds) => {
  if (value >= thresholds.danger) return "text-red-600 bg-red-50";
  if (value >= thresholds.warning) return "text-yellow-600 bg-yellow-50";
  return "text-green-600 bg-green-50";
};

// Tile background + text classes for a 0-100 reliability score (higher = healthier).
const scoreClasses = (score) => {
  const style = LEVEL_STYLES[reliabilityLevel(score ?? 0)];
  return `${style.bg} ${style.text}`;
};

export default function CarInfoPanel({ carData }) {
  if (!carData) return null;

  return (
    <div className="space-y-6">
      {/* Timestamp */}
      <div className="bg-white rounded-2xl shadow-lg p-6">
        <div className="flex items-center gap-3 mb-4">
          <Clock className="w-5 h-5 text-blue-600" />
          <h3 className="text-lg font-bold text-gray-800">
            Poslednje ažuriranje
          </h3>
        </div>
        <p className="text-sm text-gray-600">
          {formatTimestamp(carData.timestamp)}
        </p>
      </div>

      {/* Reliability */}
      <div className="bg-white rounded-2xl shadow-lg p-6">
        <div className="flex items-center gap-3 mb-4">
          <ShieldCheck className="w-5 h-5 text-emerald-600" />
          <h3 className="text-lg font-bold text-gray-800">Pouzdanost</h3>
        </div>
        <div className="space-y-3">
          <div className={`p-3 rounded-lg ${scoreClasses(carData.reliability)}`}>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <ShieldCheck className="w-4 h-4" />
                <span className="text-sm font-medium">Pouzdanost (30 dana)</span>
              </div>
              <span className="text-lg font-bold">
                {formatScore(carData.reliability)}/100
              </span>
            </div>
          </div>

          <div
            className={`p-3 rounded-lg ${scoreClasses(carData.temperature_score)}`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Flame className="w-4 h-4" />
                <span className="text-sm font-medium">
                  Skor temperature (30 dana)
                </span>
              </div>
              <span className="text-lg font-bold">
                {formatScore(carData.temperature_score)}/100
              </span>
            </div>
          </div>

          <div
            className={`p-3 rounded-lg ${checkEngineCountClasses(carData.check_engine_count)}`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <AlertTriangle className="w-4 h-4" />
                <span className="text-sm font-medium">
                  Check Engine (30 dana)
                </span>
              </div>
              <span className="text-lg font-bold">
                {carData.check_engine_count ?? 0}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Engine Stats */}
      <div className="bg-white rounded-2xl shadow-lg p-6">
        <div className="flex items-center gap-3 mb-4">
          <Activity className="w-5 h-5 text-purple-600" />
          <h3 className="text-lg font-bold text-gray-800">Status motora</h3>
        </div>
        <div className="space-y-3">
          <div
            className={`p-3 rounded-lg ${
              carData.check_engine
                ? "text-red-600 bg-red-50"
                : "text-green-600 bg-green-50"
            }`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {carData.check_engine ? (
                  <AlertTriangle className="w-4 h-4" />
                ) : (
                  <CheckCircle className="w-4 h-4" />
                )}
                <span className="text-sm font-medium">Check Engine</span>
              </div>
              <span className="text-lg font-bold">
                {carData.check_engine ? "Greška" : "OK"}
              </span>
            </div>
          </div>

          <div
            className={`p-3 rounded-lg ${getStatusColor(carData.engine_temperature, { warning: 90, danger: 100 })}`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Thermometer className="w-4 h-4" />
                <span className="text-sm font-medium">Temperatura motora</span>
              </div>
              <span className="text-lg font-bold">
                {carData.engine_temperature.toFixed(1)}°C
              </span>
            </div>
          </div>

          <div
            className={`p-3 rounded-lg ${getStatusColor(carData.engine_coolant_temp, { warning: 85, danger: 95 })}`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplet className="w-4 h-4" />
                <span className="text-sm font-medium">Rashladna tečnost</span>
              </div>
              <span className="text-lg font-bold">
                {carData.engine_coolant_temp.toFixed(1)}°C
              </span>
            </div>
          </div>

          <div className="p-3 rounded-lg bg-blue-50 text-blue-600">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Gauge className="w-4 h-4" />
                <span className="text-sm font-medium">Gas pedala</span>
              </div>
              <span className="text-lg font-bold">
                {(carData.gas_throttle * 100).toFixed(0)}%
              </span>
            </div>
          </div>

          <div
            className={`p-3 rounded-lg ${getStatusColor(100 - carData.fuel_level, { warning: 70, danger: 85 })}`}
          >
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Droplet className="w-4 h-4" />
                <span className="text-sm font-medium">Nivo goriva</span>
              </div>
              <span className="text-lg font-bold">
                {carData.fuel_level.toFixed(1)}%
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
