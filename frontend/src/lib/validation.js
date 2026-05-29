export const ROLES = ["user", "admin"];

const EMAIL_RE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const validateNewAccount = ({ email, password, role }) => {
  if (!EMAIL_RE.test(email || "")) {
    return { valid: false, error: "Unesite ispravnu email adresu" };
  }
  if (!password || password.length < 6) {
    return { valid: false, error: "Lozinka mora imati najmanje 6 karaktera" };
  }
  if (!ROLES.includes(role)) {
    return { valid: false, error: "Izaberite ispravnu rolu" };
  }
  return { valid: true, error: "" };
};
