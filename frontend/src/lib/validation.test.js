import { describe, it, expect } from "vitest";
import { validateNewAccount, ROLES } from "./validation";

describe("validateNewAccount", () => {
  it("accepts a valid account", () => {
    const result = validateNewAccount({
      email: "new@rentcar.com",
      password: "password123",
      role: "user",
    });
    expect(result.valid).toBe(true);
    expect(result.error).toBe("");
  });

  it("rejects an invalid email", () => {
    const result = validateNewAccount({
      email: "not-an-email",
      password: "password123",
      role: "user",
    });
    expect(result.valid).toBe(false);
    expect(result.error).toMatch(/email/i);
  });

  it("rejects a password shorter than 6 characters", () => {
    const result = validateNewAccount({
      email: "new@rentcar.com",
      password: "123",
      role: "user",
    });
    expect(result.valid).toBe(false);
    expect(result.error).toMatch(/lozink|6/i);
  });

  it("rejects an unknown role", () => {
    const result = validateNewAccount({
      email: "new@rentcar.com",
      password: "password123",
      role: "superuser",
    });
    expect(result.valid).toBe(false);
    expect(result.error).toMatch(/rol/i);
  });

  it("exposes the allowed roles", () => {
    expect(ROLES).toEqual(["user", "admin"]);
  });
});
