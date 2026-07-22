// Fleet helpers for the dashboard.

// filterCars turns the live cars map into a list, filtered by a case-insensitive
// substring of the car id and sorted by id so the grid order stays stable.
export const filterCars = (cars, query) => {
  const list = Object.values(cars || {});
  const q = (query || "").trim().toLowerCase();
  const filtered = q
    ? list.filter((c) => c.car_id?.toLowerCase().includes(q))
    : list;
  return filtered.sort((a, b) => a.car_id.localeCompare(b.car_id));
};
