// The backend returns errors as { "error": "..." }. parseApiError extracts a
// human-readable message, tolerating the older "message" shape and missing data.
export const parseApiError = (data, fallback) =>
  data?.error || data?.message || fallback;
