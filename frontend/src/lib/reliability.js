// Reliability and its sub-scores are 0-100 (higher = healthier), computed by the
// backend over the trailing 30 days. These helpers map a score to a dashboard
// severity band and format it for display.

export const reliabilityLevel = (score) =>
  score >= 80 ? "ok" : score >= 50 ? "warn" : "critical";

// Round a 0-100 score for display; a missing value renders as 0.
export const formatScore = (score) => String(Math.round(score ?? 0));

// scoreColor maps a 0-100 score to a hex color for gauges/bars, matching the
// three severity bands (green / amber / red).
export const scoreColor = (score) => {
  const level = reliabilityLevel(score ?? 0);
  if (level === "ok") return "#16a34a";
  if (level === "warn") return "#d97706";
  return "#dc2626";
};

// checkEngineCountClasses colors the check-engine tile by episode count rather
// than by score: healthy (0) is green, a single light is yellow, and each
// additional light escalates the shade (orange, then red).
export const checkEngineCountClasses = (count) => {
  if (!count) return "bg-green-50 text-green-600";
  if (count === 1) return "bg-yellow-50 text-yellow-700";
  if (count === 2) return "bg-orange-50 text-orange-600";
  return "bg-red-50 text-red-600";
};
