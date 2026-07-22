// Account approval status, mirroring the backend auth statuses. Self-registered
// accounts are pending until an admin approves them.

export const STATUS_PENDING = "pending";
export const STATUS_APPROVED = "approved";

export const statusLabel = (status) => {
  if (status === STATUS_APPROVED) return "Odobren";
  if (status === STATUS_PENDING) return "Na čekanju";
  return "Nepoznato";
};

export const isPending = (status) => status === STATUS_PENDING;
