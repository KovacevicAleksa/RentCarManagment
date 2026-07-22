import { useState, useEffect } from "react";
import {
  ShieldCheck,
  UserPlus,
  Mail,
  Lock,
  ArrowLeft,
  Users,
  KeyRound,
  LogOut,
  Trash2,
  Bell,
  Send,
  UserCheck,
  Clock,
} from "lucide-react";
import { validateNewAccount, ROLES } from "../../lib/validation";
import { parseApiError } from "../../lib/api";
import { statusLabel, isPending } from "../../lib/accountStatus";
import { sendNotification } from "../../lib/notifications";
import { useWebSocket } from "../../contexts/WebSocketContext";
import NotificationBell from "../notifications/NotificationBell";
import NotificationsPage from "../notifications/NotificationsPage";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";

const ROLE_LABELS = { user: "Korisnik", admin: "Administrator" };

export default function AdminPanel({ onBack, onLogout, userEmail, currentUserId }) {
  const { unreadCount, ringKey } = useWebSocket();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("user");
  const [message, setMessage] = useState(null); // { type: "success" | "error", text }
  const [submitting, setSubmitting] = useState(false);
  const [users, setUsers] = useState([]);
  const [loadingUsers, setLoadingUsers] = useState(true);
  const [resetResult, setResetResult] = useState(null); // { email, password }
  const [resettingId, setResettingId] = useState(null);
  const [deletingId, setDeletingId] = useState(null);
  const [approvingId, setApprovingId] = useState(null);
  const [showNotifications, setShowNotifications] = useState(false);

  // "Send notification" form
  const [notifTitle, setNotifTitle] = useState("");
  const [notifMessage, setNotifMessage] = useState("");
  const [notifType, setNotifType] = useState("info");
  const [sendingNotif, setSendingNotif] = useState(false);
  const [notifFeedback, setNotifFeedback] = useState(null); // { type, text }

  const fetchUsers = async () => {
    setLoadingUsers(true);
    try {
      const res = await fetch(`${API_URL}/admin/users`, {
        credentials: "include",
      });
      if (!res.ok) throw new Error("Neuspešno učitavanje naloga");
      const data = await res.json();
      setUsers(data.users || []);
    } catch (err) {
      setMessage({ type: "error", text: err.message });
    } finally {
      setLoadingUsers(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const resetPassword = async (u) => {
    if (!window.confirm(`Resetovati lozinku za ${u.email}?`)) return;
    setResetResult(null);
    setResettingId(u.id);
    try {
      const res = await fetch(`${API_URL}/admin/users/${u.id}/reset-password`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (!res.ok) throw new Error(parseApiError(data, "Reset lozinke nije uspeo"));
      setResetResult({ email: u.email, password: data.temporary_password });
    } catch (err) {
      setMessage({ type: "error", text: err.message });
    } finally {
      setResettingId(null);
    }
  };

  const approveUser = async (u) => {
    setMessage(null);
    setApprovingId(u.id);
    try {
      const res = await fetch(`${API_URL}/admin/users/${u.id}/approve`, {
        method: "POST",
        credentials: "include",
      });
      const data = await res.json();
      if (!res.ok) throw new Error(parseApiError(data, "Odobravanje nije uspelo"));
      setMessage({ type: "success", text: `Nalog ${u.email} je odobren` });
      fetchUsers();
    } catch (err) {
      setMessage({ type: "error", text: err.message });
    } finally {
      setApprovingId(null);
    }
  };

  const deleteUser = async (u) => {
    if (!window.confirm(`Obrisati nalog ${u.email}? Ova akcija se ne može opozvati.`)) return;
    setMessage(null);
    setDeletingId(u.id);
    try {
      const res = await fetch(`${API_URL}/admin/users/${u.id}`, {
        method: "DELETE",
        credentials: "include",
      });
      const data = await res.json();
      if (!res.ok) throw new Error(parseApiError(data, "Brisanje nije uspelo"));
      setMessage({ type: "success", text: `Nalog ${u.email} je obrisan` });
      fetchUsers();
    } catch (err) {
      setMessage({ type: "error", text: err.message });
    } finally {
      setDeletingId(null);
    }
  };

  const handleCreate = async () => {
    setMessage(null);

    const { valid, error } = validateNewAccount({ email, password, role });
    if (!valid) {
      setMessage({ type: "error", text: error });
      return;
    }

    setSubmitting(true);
    try {
      const res = await fetch(`${API_URL}/admin/users`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password, role }),
        credentials: "include",
      });
      const data = await res.json();
      if (!res.ok) {
        throw new Error(parseApiError(data, "Kreiranje naloga nije uspelo"));
      }
      setMessage({ type: "success", text: `Nalog ${email} je kreiran` });
      setEmail("");
      setPassword("");
      setRole("user");
      fetchUsers();
    } catch (err) {
      setMessage({ type: "error", text: err.message });
    } finally {
      setSubmitting(false);
    }
  };

  const handleSendNotification = async () => {
    setNotifFeedback(null);
    const title = notifTitle.trim();
    const text = notifMessage.trim();
    if (!title || !text) {
      setNotifFeedback({ type: "error", text: "Naslov i poruka su obavezni" });
      return;
    }

    setSendingNotif(true);
    try {
      await sendNotification({ title, message: text, type: notifType });
      setNotifFeedback({ type: "success", text: "Obaveštenje je poslato svim korisnicima" });
      setNotifTitle("");
      setNotifMessage("");
      setNotifType("info");
    } catch (err) {
      setNotifFeedback({ type: "error", text: err.message });
    } finally {
      setSendingNotif(false);
    }
  };

  if (showNotifications) {
    return <NotificationsPage onBack={() => setShowNotifications(false)} />;
  }

  const pendingCount = users.filter((u) => isPending(u.status)).length;
  // Show pending accounts first so an admin can approve them at a glance.
  const sortedUsers = [...users].sort(
    (a, b) => Number(isPending(b.status)) - Number(isPending(a.status)),
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <header className="bg-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-md">
                <ShieldCheck className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-800">
                Administracija
              </h1>
            </div>
            <div className="flex items-center gap-3">
              <NotificationBell
                count={unreadCount}
                ringKey={ringKey}
                onClick={() => setShowNotifications(true)}
              />
              {userEmail && (
                <span className="hidden sm:block text-sm text-gray-600">
                  {userEmail}
                </span>
              )}
              {onLogout ? (
                <button
                  onClick={onLogout}
                  className="flex items-center gap-2 px-4 py-2 text-red-600 hover:bg-red-50 rounded-xl transition-all font-medium"
                >
                  <LogOut className="w-4 h-4" />
                  Odjavi se
                </button>
              ) : (
                <button
                  onClick={onBack}
                  className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-xl transition-all"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Nazad
                </button>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        <section className="bg-white rounded-2xl shadow-lg p-6">
          <div className="flex items-center gap-2 mb-6">
            <UserPlus className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-bold text-gray-800">Novi nalog</h2>
          </div>

          <div className="space-y-5">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Email adresa
              </label>
              <div className="relative">
                <Mail className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="email"
                  placeholder="ime@primer.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Lozinka
              </label>
              <div className="relative">
                <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
                <input
                  type="password"
                  placeholder="Najmanje 6 karaktera"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                />
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Rola
              </label>
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-white"
              >
                {ROLES.map((r) => (
                  <option key={r} value={r}>
                    {ROLE_LABELS[r] || r}
                  </option>
                ))}
              </select>
            </div>

            <button
              onClick={handleCreate}
              disabled={submitting}
              className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transform hover:scale-[1.02] transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
            >
              {submitting ? "Kreiranje..." : "Kreiraj nalog"}
            </button>

            {message && (
              <div
                className={`p-4 rounded-xl text-sm ${
                  message.type === "success"
                    ? "bg-green-50 text-green-700 border border-green-200"
                    : "bg-red-50 text-red-700 border border-red-200"
                }`}
              >
                {message.text}
              </div>
            )}
          </div>
        </section>

        <section className="bg-white rounded-2xl shadow-lg p-6">
          <div className="flex items-center gap-2 mb-6">
            <Users className="w-5 h-5 text-purple-600" />
            <h2 className="text-lg font-bold text-gray-800">
              Nalozi ({users.length})
            </h2>
            {pendingCount > 0 && (
              <span className="flex items-center gap-1 text-xs font-semibold px-2.5 py-1 rounded-full bg-amber-100 text-amber-700">
                <Clock className="w-3.5 h-3.5" />
                {pendingCount} na čekanju
              </span>
            )}
          </div>

          {resetResult && (
            <div className="mb-4 p-4 rounded-xl bg-amber-50 border border-amber-200">
              <p className="text-sm text-amber-800 mb-2">
                Privremena lozinka za <strong>{resetResult.email}</strong> —
                prikazuje se samo jednom, prosledi je korisniku:
              </p>
              <div className="flex items-center justify-between gap-3">
                <code
                  data-testid="temp-password"
                  className="flex-1 px-3 py-2 bg-white border border-amber-300 rounded-lg font-mono text-sm text-gray-800 select-all"
                >
                  {resetResult.password}
                </code>
                <button
                  onClick={() => setResetResult(null)}
                  className="text-xs text-amber-700 hover:text-amber-900 font-medium"
                >
                  Zatvori
                </button>
              </div>
            </div>
          )}

          {loadingUsers ? (
            <p className="text-gray-500 text-sm">Učitavanje...</p>
          ) : users.length === 0 ? (
            <p className="text-gray-500 text-sm">Nema naloga</p>
          ) : (
            <ul className="divide-y divide-gray-100">
              {sortedUsers.map((u) => (
                <li
                  key={u.id}
                  className={`flex items-center justify-between gap-3 py-3 px-2 rounded-lg ${
                    isPending(u.status) ? "bg-amber-50" : ""
                  }`}
                >
                  <span className="text-sm text-gray-700 truncate flex-1">
                    {u.email}
                  </span>
                  <span
                    className={`hidden sm:inline text-xs font-semibold px-2.5 py-1 rounded-full ${
                      isPending(u.status)
                        ? "bg-amber-100 text-amber-700"
                        : "bg-green-100 text-green-700"
                    }`}
                  >
                    {statusLabel(u.status)}
                  </span>
                  <span
                    className={`text-xs font-semibold px-2.5 py-1 rounded-full ${
                      u.role === "admin"
                        ? "bg-purple-100 text-purple-700"
                        : "bg-blue-100 text-blue-700"
                    }`}
                  >
                    {ROLE_LABELS[u.role] || u.role}
                  </span>
                  {isPending(u.status) && (
                    <button
                      onClick={() => approveUser(u)}
                      disabled={approvingId === u.id}
                      title="Odobri nalog"
                      className="flex items-center gap-1 text-xs text-green-700 hover:text-white hover:bg-green-600 bg-green-100 px-2 py-1 rounded-lg transition-all disabled:opacity-50 font-medium"
                    >
                      <UserCheck className="w-3.5 h-3.5" />
                      {approvingId === u.id ? "..." : "Odobri"}
                    </button>
                  )}
                  <button
                    onClick={() => resetPassword(u)}
                    disabled={resettingId === u.id}
                    title="Resetuj lozinku"
                    className="flex items-center gap-1 text-xs text-gray-500 hover:text-blue-600 hover:bg-blue-50 px-2 py-1 rounded-lg transition-all disabled:opacity-50"
                  >
                    <KeyRound className="w-3.5 h-3.5" />
                    {resettingId === u.id ? "..." : "Reset"}
                  </button>
                  {u.id !== currentUserId && (
                    <button
                      onClick={() => deleteUser(u)}
                      disabled={deletingId === u.id}
                      title="Obriši nalog"
                      className="flex items-center gap-1 text-xs text-gray-500 hover:text-red-600 hover:bg-red-50 px-2 py-1 rounded-lg transition-all disabled:opacity-50"
                    >
                      <Trash2 className="w-3.5 h-3.5" />
                      {deletingId === u.id ? "..." : "Obriši"}
                    </button>
                  )}
                </li>
              ))}
            </ul>
          )}
        </section>

        <section className="bg-white rounded-2xl shadow-lg p-6">
          <div className="flex items-center gap-2 mb-6">
            <Bell className="w-5 h-5 text-blue-600" />
            <h2 className="text-lg font-bold text-gray-800">
              Pošalji obaveštenje svim korisnicima
            </h2>
          </div>

          <div className="space-y-5">
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="md:col-span-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Naslov
                </label>
                <input
                  type="text"
                  placeholder="npr. Planirano održavanje"
                  value={notifTitle}
                  onChange={(e) => setNotifTitle(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Tip
                </label>
                <select
                  value={notifType}
                  onChange={(e) => setNotifType(e.target.value)}
                  className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all bg-white"
                >
                  <option value="info">Informacija</option>
                  <option value="alert">Upozorenje</option>
                </select>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Poruka
              </label>
              <textarea
                rows={3}
                placeholder="Tekst obaveštenja..."
                value={notifMessage}
                onChange={(e) => setNotifMessage(e.target.value)}
                className="w-full px-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all resize-none"
              />
            </div>

            <button
              onClick={handleSendNotification}
              disabled={sendingNotif}
              className="flex items-center justify-center gap-2 w-full md:w-auto px-6 bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transform hover:scale-[1.02] transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
            >
              <Send className="w-4 h-4" />
              {sendingNotif ? "Slanje..." : "Pošalji obaveštenje"}
            </button>

            {notifFeedback && (
              <div
                className={`p-4 rounded-xl text-sm ${
                  notifFeedback.type === "success"
                    ? "bg-green-50 text-green-700 border border-green-200"
                    : "bg-red-50 text-red-700 border border-red-200"
                }`}
              >
                {notifFeedback.text}
              </div>
            )}
          </div>
        </section>
      </main>
    </div>
  );
}
