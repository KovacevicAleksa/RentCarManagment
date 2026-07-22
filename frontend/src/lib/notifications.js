const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

export const fetchNotifications = async () => {
  const res = await fetch(`${API_URL}/notifications`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Neuspešno učitavanje obaveštenja");
  const data = await res.json();
  return data.notifications || [];
};

export const fetchUnreadCount = async () => {
  const res = await fetch(`${API_URL}/notifications/unread-count`, {
    credentials: "include",
  });
  if (!res.ok) throw new Error("Neuspešno učitavanje broja obaveštenja");
  const data = await res.json();
  return data.unread || 0;
};

export const markNotificationsRead = async () => {
  const res = await fetch(`${API_URL}/notifications/read`, {
    method: "POST",
    credentials: "include",
  });
  if (!res.ok) throw new Error("Neuspešno označavanje obaveštenja");
};

export const sendNotification = async ({ title, message, type }) => {
  const res = await fetch(`${API_URL}/admin/notifications`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ title, message, type }),
    credentials: "include",
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data?.error || "Slanje obaveštenja nije uspelo");
  }
  return data.notification;
};

// notificationStyle maps a notification type to its visual treatment. Pure so
// it can be unit-tested and reused across the bell and the notifications page.
export const notificationStyle = (type) => {
  if (type === "alert") {
    return {
      label: "Upozorenje",
      dot: "bg-red-500",
      badge: "bg-red-100 text-red-700",
      icon: "text-red-600",
      card: "border-red-200 bg-red-50",
    };
  }
  return {
    label: "Informacija",
    dot: "bg-blue-500",
    badge: "bg-blue-100 text-blue-700",
    icon: "text-blue-600",
    card: "border-blue-200 bg-blue-50",
  };
};
