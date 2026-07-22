import { describe, it, expect } from "vitest";
import { toChartRows, faultTotal } from "./events";

describe("toChartRows", () => {
  it("maps daily buckets to DD.MM chart rows", () => {
    const rows = toChartRows([
      { date: "2026-07-21", overheat: 1, check_engine: 0 },
      { date: "2026-07-22", overheat: 2, check_engine: 3 },
    ]);
    expect(rows).toEqual([
      { label: "21.07", overheat: 1, checkEngine: 0 },
      { label: "22.07", overheat: 2, checkEngine: 3 },
    ]);
  });

  it("handles missing/empty input", () => {
    expect(toChartRows()).toEqual([]);
    expect(toChartRows([])).toEqual([]);
  });
});

describe("faultTotal", () => {
  it("sums overheat and check-engine totals", () => {
    expect(faultTotal({ overheat_total: 2, check_engine_total: 3 })).toBe(5);
  });
  it("treats missing stats as 0", () => {
    expect(faultTotal(undefined)).toBe(0);
    expect(faultTotal({})).toBe(0);
  });
});
