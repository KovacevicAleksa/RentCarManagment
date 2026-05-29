const STORAGE_KEY = "rentcar.preferences";

export const DEFAULT_PREFS = {
  engineWarn: 90,
  engineCritical: 100,
  coolantWarn: 85,
  coolantCritical: 95,
  fuelWarn: 30,
  fuelCritical: 15,
};

const TEMP_KEYS = ["engineWarn", "engineCritical", "coolantWarn", "coolantCritical"];
const FUEL_KEYS = ["fuelWarn", "fuelCritical"];

export const validateThresholds = (p) => {
  for (const k of [...TEMP_KEYS, ...FUEL_KEYS]) {
    if (typeof p[k] !== "number" || Number.isNaN(p[k])) {
      return { valid: false, error: `Vrednost "${k}" mora biti broj` };
    }
  }
  for (const k of TEMP_KEYS) {
    if (p[k] < 0 || p[k] > 200) {
      return { valid: false, error: "Temperature moraju biti između 0 i 200°C" };
    }
  }
  for (const k of FUEL_KEYS) {
    if (p[k] < 0 || p[k] > 100) {
      return { valid: false, error: "Pragovi goriva moraju biti između 0 i 100%" };
    }
  }
  if (p.engineWarn >= p.engineCritical) {
    return { valid: false, error: "Prag upozorenja motora mora biti manji od kritičnog" };
  }
  if (p.coolantWarn >= p.coolantCritical) {
    return { valid: false, error: "Prag upozorenja rashladne mora biti manji od kritičnog" };
  }
  if (p.fuelCritical >= p.fuelWarn) {
    return { valid: false, error: "Kritični prag goriva mora biti manji od praga upozorenja" };
  }
  return { valid: true, error: "" };
};

export const loadPreferences = () => {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return { ...DEFAULT_PREFS };
    const parsed = JSON.parse(raw);
    return { ...DEFAULT_PREFS, ...parsed };
  } catch {
    return { ...DEFAULT_PREFS };
  }
};

export const savePreferences = (prefs) => {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(prefs));
};
