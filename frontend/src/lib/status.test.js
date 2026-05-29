import { describe, it, expect } from "vitest";
import { engineTempLevel, coolantLevel, fuelLevel } from "./status";

const T = { engineWarn: 90, engineCritical: 100, coolantWarn: 85, coolantCritical: 95 };

describe("engineTempLevel", () => {
  it("ok below warn", () => expect(engineTempLevel(80, T)).toBe("ok"));
  it("warn at threshold", () => expect(engineTempLevel(90, T)).toBe("warn"));
  it("warn between thresholds", () => expect(engineTempLevel(95, T)).toBe("warn"));
  it("critical at threshold", () => expect(engineTempLevel(100, T)).toBe("critical"));
  it("critical above", () => expect(engineTempLevel(120, T)).toBe("critical"));
});

describe("coolantLevel", () => {
  it("ok", () => expect(coolantLevel(80, T)).toBe("ok"));
  it("warn", () => expect(coolantLevel(85, T)).toBe("warn"));
  it("critical", () => expect(coolantLevel(95, T)).toBe("critical"));
});

describe("fuelLevel (inverted: low is bad)", () => {
  const F = { fuelWarn: 30, fuelCritical: 15 };
  it("ok when full", () => expect(fuelLevel(80, F)).toBe("ok"));
  it("warn at threshold", () => expect(fuelLevel(30, F)).toBe("warn"));
  it("warn between", () => expect(fuelLevel(20, F)).toBe("warn"));
  it("critical at threshold", () => expect(fuelLevel(15, F)).toBe("critical"));
  it("critical when empty", () => expect(fuelLevel(2, F)).toBe("critical"));
});
