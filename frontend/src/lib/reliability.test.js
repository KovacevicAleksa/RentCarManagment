import { describe, it, expect } from "vitest";
import {
  reliabilityLevel,
  formatScore,
  scoreColor,
  checkEngineCountClasses,
} from "./reliability";

describe("reliabilityLevel (higher score = healthier)", () => {
  it("ok at the top of the range", () => {
    expect(reliabilityLevel(100)).toBe("ok");
    expect(reliabilityLevel(80)).toBe("ok");
  });
  it("warn in the mid band", () => {
    expect(reliabilityLevel(79)).toBe("warn");
    expect(reliabilityLevel(50)).toBe("warn");
  });
  it("critical when low", () => {
    expect(reliabilityLevel(49)).toBe("critical");
    expect(reliabilityLevel(0)).toBe("critical");
  });
});

describe("formatScore", () => {
  it("rounds to the nearest integer", () => {
    expect(formatScore(88.4)).toBe("88");
    expect(formatScore(88.6)).toBe("89");
  });
  it("treats a missing score as 0", () => {
    expect(formatScore(undefined)).toBe("0");
    expect(formatScore(null)).toBe("0");
  });
});

describe("scoreColor", () => {
  it("maps score bands to hex colors", () => {
    expect(scoreColor(90)).toBe("#16a34a"); // green
    expect(scoreColor(60)).toBe("#d97706"); // amber
    expect(scoreColor(30)).toBe("#dc2626"); // red
  });
});

describe("checkEngineCountClasses (progressive by count)", () => {
  it("is green when there are no check-engine lights", () => {
    expect(checkEngineCountClasses(0)).toContain("green");
    expect(checkEngineCountClasses(undefined)).toContain("green");
  });
  it("is yellow at one, and intensifies above one", () => {
    expect(checkEngineCountClasses(1)).toContain("yellow");
    expect(checkEngineCountClasses(2)).toContain("orange");
    expect(checkEngineCountClasses(3)).toContain("red");
    expect(checkEngineCountClasses(9)).toContain("red");
  });
});
