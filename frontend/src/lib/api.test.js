import { describe, it, expect } from "vitest";
import { parseApiError } from "./api";

describe("parseApiError", () => {
  it("prefers the backend 'error' field", () => {
    expect(parseApiError({ error: "user already exists" }, "fallback")).toBe(
      "user already exists",
    );
  });

  it("falls back to 'message' when 'error' is absent", () => {
    expect(parseApiError({ message: "something" }, "fallback")).toBe(
      "something",
    );
  });

  it("uses the fallback when neither field is present", () => {
    expect(parseApiError({}, "fallback")).toBe("fallback");
  });

  it("uses the fallback for null/undefined data", () => {
    expect(parseApiError(null, "fallback")).toBe("fallback");
    expect(parseApiError(undefined, "fallback")).toBe("fallback");
  });
});
