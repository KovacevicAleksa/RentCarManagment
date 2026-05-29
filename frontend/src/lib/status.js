// Severity helpers shared by the dashboard. Each returns "ok" | "warn" | "critical".

export const engineTempLevel = (temp, { engineWarn, engineCritical }) =>
  temp >= engineCritical ? "critical" : temp >= engineWarn ? "warn" : "ok";

export const coolantLevel = (temp, { coolantWarn, coolantCritical }) =>
  temp >= coolantCritical ? "critical" : temp >= coolantWarn ? "warn" : "ok";

// Fuel is inverted: a low level is the bad case.
export const fuelLevel = (fuel, { fuelWarn, fuelCritical }) =>
  fuel <= fuelCritical ? "critical" : fuel <= fuelWarn ? "warn" : "ok";

// Tailwind classes per severity, used for the metric tiles.
export const LEVEL_STYLES = {
  ok: { bg: "bg-green-50", text: "text-green-600", icon: "text-green-600" },
  warn: { bg: "bg-yellow-50", text: "text-yellow-600", icon: "text-yellow-600" },
  critical: { bg: "bg-red-50", text: "text-red-600", icon: "text-red-600" },
};
