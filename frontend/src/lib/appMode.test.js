import { describe, it, expect } from "vitest";
import { resolveAppMode } from "./appMode";

describe("resolveAppMode", () => {
  it("returns 'admin' when mode is 'admin'", () => {
    expect(resolveAppMode("admin")).toBe("admin");
  });
  it("is case-insensitive", () => {
    expect(resolveAppMode("ADMIN")).toBe("admin");
  });
  it("defaults to 'user' for anything else", () => {
    expect(resolveAppMode("user")).toBe("user");
    expect(resolveAppMode("")).toBe("user");
    expect(resolveAppMode(undefined)).toBe("user");
    expect(resolveAppMode("something")).toBe("user");
  });
});
