import { describe, it, expect } from "vitest";
import { notificationStyle } from "./notifications";

describe("notificationStyle", () => {
  it("uses red treatment for alerts", () => {
    const s = notificationStyle("alert");
    expect(s.label).toBe("Upozorenje");
    expect(s.dot).toContain("red");
  });

  it("falls back to info treatment for anything else", () => {
    expect(notificationStyle("info").label).toBe("Informacija");
    expect(notificationStyle(undefined).label).toBe("Informacija");
    expect(notificationStyle("whatever").dot).toContain("blue");
  });
});
