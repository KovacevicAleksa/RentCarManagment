// The app is built from one codebase but deployed as two containers.
// VITE_APP_MODE selects which experience to render.
export const resolveAppMode = (mode) =>
  String(mode || "").toLowerCase() === "admin" ? "admin" : "user";
