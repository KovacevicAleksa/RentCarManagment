// Helpers for the per-car fault history (overheat + check-engine episodes)
// returned by GET /carstats/car/:carID/events.

// toChartRows maps the backend's daily buckets ({date: "YYYY-MM-DD", overheat,
// check_engine}) into recharts rows with a short DD.MM label.
export const toChartRows = (buckets = []) =>
  buckets.map((b) => ({
    label: `${b.date.slice(8, 10)}.${b.date.slice(5, 7)}`,
    overheat: b.overheat,
    checkEngine: b.check_engine,
  }));

// faultTotal is the combined number of fault episodes in the window.
export const faultTotal = (stats) =>
  (stats?.overheat_total ?? 0) + (stats?.check_engine_total ?? 0);
