import { describe, it, expect } from "vitest";
import { DEFAULT_PREFS, validateThresholds } from "./preferences";

describe("validateThresholds", () => {
  it("accepts the defaults", () => {
    expect(validateThresholds(DEFAULT_PREFS).valid).toBe(true);
  });

  it("rejects engine warn >= critical", () => {
    const r = validateThresholds({ ...DEFAULT_PREFS, engineWarn: 110 });
    expect(r.valid).toBe(false);
    expect(r.error).toMatch(/motor/i);
  });

  it("rejects fuel critical >= warn (inverted)", () => {
    const r = validateThresholds({ ...DEFAULT_PREFS, fuelCritical: 40 });
    expect(r.valid).toBe(false);
    expect(r.error).toMatch(/goriv/i);
  });

  it("rejects non-numeric values", () => {
    const r = validateThresholds({ ...DEFAULT_PREFS, engineWarn: NaN });
    expect(r.valid).toBe(false);
  });

  it("rejects out-of-range temperatures", () => {
    expect(validateThresholds({ ...DEFAULT_PREFS, engineCritical: 500 }).valid).toBe(false);
  });

  it("rejects out-of-range fuel", () => {
    expect(validateThresholds({ ...DEFAULT_PREFS, fuelWarn: 150 }).valid).toBe(false);
  });
});
