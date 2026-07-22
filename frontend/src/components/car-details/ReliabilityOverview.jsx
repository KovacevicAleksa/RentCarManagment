import { ShieldCheck, Flame, AlertTriangle } from "lucide-react";
import {
  formatScore,
  scoreColor,
  checkEngineCountClasses,
} from "../../lib/reliability";

// Radial 0-100 gauge drawn as an SVG ring.
function Gauge({ score }) {
  const value = Math.max(0, Math.min(100, score ?? 0));
  const r = 60;
  const c = 2 * Math.PI * r;
  const offset = c * (1 - value / 100);
  const color = scoreColor(value);

  return (
    <svg viewBox="0 0 140 140" className="w-40 h-40 shrink-0">
      <circle cx="70" cy="70" r={r} fill="none" stroke="#e5e7eb" strokeWidth="12" />
      <circle
        cx="70"
        cy="70"
        r={r}
        fill="none"
        stroke={color}
        strokeWidth="12"
        strokeLinecap="round"
        strokeDasharray={c}
        strokeDashoffset={offset}
        transform="rotate(-90 70 70)"
      />
      <text
        x="70"
        y="66"
        textAnchor="middle"
        className="font-extrabold"
        style={{ fontSize: "30px", fill: color }}
      >
        {formatScore(value)}
      </text>
      <text x="70" y="90" textAnchor="middle" style={{ fontSize: "13px", fill: "#6b7280" }}>
        / 100
      </text>
    </svg>
  );
}

// Labeled horizontal 0-100 sub-score bar.
function ScoreBar({ icon, label, score }) {
  const Icon = icon;
  const value = Math.max(0, Math.min(100, score ?? 0));
  const color = scoreColor(value);
  return (
    <div>
      <div className="flex items-center justify-between mb-1.5">
        <div className="flex items-center gap-2 text-sm font-medium text-gray-700">
          <Icon className="w-4 h-4" style={{ color }} />
          {label}
        </div>
        <span className="text-sm font-bold" style={{ color }}>
          {formatScore(value)}/100
        </span>
      </div>
      <div className="h-2.5 w-full rounded-full bg-gray-100 overflow-hidden">
        <div
          className="h-full rounded-full transition-all"
          style={{ width: `${value}%`, backgroundColor: color }}
        />
      </div>
    </div>
  );
}

export default function ReliabilityOverview({ carData }) {
  if (!carData) return null;

  const overheatCount = carData.overheat_count ?? 0;
  const checkEngineCount = carData.check_engine_count ?? 0;

  return (
    <div className="bg-white rounded-2xl shadow-lg p-6">
      <div className="flex items-center gap-3 mb-6">
        <ShieldCheck className="w-5 h-5 text-emerald-600" />
        <h3 className="text-lg font-bold text-gray-800">Pregled pouzdanosti</h3>
        <span className="text-xs text-gray-500 ml-auto">poslednjih 30 dana</span>
      </div>

      <div className="flex flex-col sm:flex-row items-center gap-6">
        <div className="flex flex-col items-center">
          <Gauge score={carData.reliability} />
          <p className="text-sm font-semibold text-gray-700 mt-1">Ukupna pouzdanost</p>
        </div>

        <div className="flex-1 w-full space-y-4">
          <ScoreBar icon={Flame} label="Temperatura" score={carData.temperature_score} />
          <ScoreBar icon={AlertTriangle} label="Check Engine" score={carData.check_engine_score} />

          <div className="flex flex-wrap gap-2 pt-1">
            <span className="inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-full bg-orange-50 text-orange-600">
              <Flame className="w-3.5 h-3.5" />
              {overheatCount} pregrevanja
            </span>
            <span
              className={`inline-flex items-center gap-1.5 text-xs font-semibold px-3 py-1.5 rounded-full ${checkEngineCountClasses(checkEngineCount)}`}
            >
              <AlertTriangle className="w-3.5 h-3.5" />
              {checkEngineCount} check engine
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
