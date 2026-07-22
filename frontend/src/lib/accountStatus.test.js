import { describe, it, expect } from "vitest";
import { statusLabel, isPending, STATUS_PENDING, STATUS_APPROVED } from "./accountStatus";

describe("statusLabel", () => {
  it("labels approved and pending in Serbian", () => {
    expect(statusLabel(STATUS_APPROVED)).toBe("Odobren");
    expect(statusLabel(STATUS_PENDING)).toBe("Na čekanju");
  });
  it("falls back to a placeholder for unknown/missing status", () => {
    expect(statusLabel(undefined)).toBe("Nepoznato");
    expect(statusLabel("")).toBe("Nepoznato");
  });
});

describe("isPending", () => {
  it("is true only for the pending status", () => {
    expect(isPending(STATUS_PENDING)).toBe(true);
    expect(isPending(STATUS_APPROVED)).toBe(false);
    expect(isPending(undefined)).toBe(false);
  });
});
