import { useState } from "react";
import {
  Settings as SettingsIcon,
  ArrowLeft,
  User,
  Lock,
  Mail,
  SlidersHorizontal,
  Server,
  ShieldCheck,
  LogOut,
  RotateCcw,
  Wifi,
  WifiOff,
} from "lucide-react";
import { useWebSocket } from "../../contexts/WebSocketContext";
import { parseApiError } from "../../lib/api";
import { DEFAULT_PREFS, validateThresholds, savePreferences } from "../../lib/preferences";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";
const APP_VERSION = "1.0.0";
const ROLE_LABELS = { user: "Korisnik", admin: "Administrator" };

const SECTIONS = [
  { id: "profile", label: "Profil", icon: User },
  { id: "security", label: "Bezbednost", icon: Lock },
  { id: "display", label: "Prikaz", icon: SlidersHorizontal },
  { id: "system", label: "Sistem", icon: Server },
];

function Toast({ msg }) {
  if (!msg) return null;
  return (
    <div
      className={`p-3 rounded-xl text-sm ${
        msg.type === "success"
          ? "bg-green-50 text-green-700 border border-green-200"
          : "bg-red-50 text-red-700 border border-red-200"
      }`}
    >
      {msg.text}
    </div>
  );
}

const Field = ({ label, ...props }) => (
  <div>
    <label className="block text-sm font-medium text-gray-700 mb-2">{label}</label>
    <input
      className="w-full px-4 py-2.5 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
      {...props}
    />
  </div>
);

export default function Settings({
  userInfo,
  prefs,
  onPrefsChange,
  onProfileChanged,
  onOpenAdmin,
  onLogout,
  onBack,
}) {
  const { wsConnected } = useWebSocket();
  const [active, setActive] = useState("profile");

  // password
  const [pwCurrent, setPwCurrent] = useState("");
  const [pwNew, setPwNew] = useState("");
  const [pwConfirm, setPwConfirm] = useState("");
  const [pwMsg, setPwMsg] = useState(null);
  const [pwBusy, setPwBusy] = useState(false);

  // email
  const [newEmail, setNewEmail] = useState("");
  const [emailPw, setEmailPw] = useState("");
  const [emailMsg, setEmailMsg] = useState(null);
  const [emailBusy, setEmailBusy] = useState(false);

  // thresholds (local draft)
  const [draft, setDraft] = useState(prefs);
  const [prefMsg, setPrefMsg] = useState(null);

  const changePassword = async () => {
    setPwMsg(null);
    if (pwNew.length < 6) {
      setPwMsg({ type: "error", text: "Nova lozinka mora imati najmanje 6 karaktera" });
      return;
    }
    if (pwNew !== pwConfirm) {
      setPwMsg({ type: "error", text: "Nove lozinke se ne poklapaju" });
      return;
    }
    setPwBusy(true);
    try {
      const res = await fetch(`${API_URL}/auth/password`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ current_password: pwCurrent, new_password: pwNew }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(parseApiError(data, "Promena lozinke nije uspela"));
      setPwMsg({ type: "success", text: "Lozinka je uspešno promenjena" });
      setPwCurrent("");
      setPwNew("");
      setPwConfirm("");
    } catch (e) {
      setPwMsg({ type: "error", text: e.message });
    } finally {
      setPwBusy(false);
    }
  };

  const changeEmail = async () => {
    setEmailMsg(null);
    setEmailBusy(true);
    try {
      const res = await fetch(`${API_URL}/auth/email`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        credentials: "include",
        body: JSON.stringify({ new_email: newEmail, current_password: emailPw }),
      });
      const data = await res.json();
      if (!res.ok) throw new Error(parseApiError(data, "Promena email-a nije uspela"));
      setEmailMsg({ type: "success", text: `Email promenjen na ${data.email}` });
      setNewEmail("");
      setEmailPw("");
      onProfileChanged?.();
    } catch (e) {
      setEmailMsg({ type: "error", text: e.message });
    } finally {
      setEmailBusy(false);
    }
  };

  const setNum = (key) => (e) =>
    setDraft((d) => ({ ...d, [key]: e.target.value === "" ? "" : Number(e.target.value) }));

  const savePrefs = () => {
    setPrefMsg(null);
    const { valid, error } = validateThresholds(draft);
    if (!valid) {
      setPrefMsg({ type: "error", text: error });
      return;
    }
    savePreferences(draft);
    onPrefsChange(draft);
    setPrefMsg({ type: "success", text: "Podešavanja prikaza su sačuvana" });
  };

  const resetPrefs = () => {
    setDraft(DEFAULT_PREFS);
    savePreferences(DEFAULT_PREFS);
    onPrefsChange(DEFAULT_PREFS);
    setPrefMsg({ type: "success", text: "Vraćeno na podrazumevane vrednosti" });
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <header className="bg-white shadow-md">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-3">
              <div className="flex items-center justify-center w-10 h-10 bg-gradient-to-br from-blue-600 to-purple-600 rounded-xl shadow-md">
                <SettingsIcon className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-800">Podešavanja</h1>
            </div>
            <button
              onClick={onBack}
              className="flex items-center gap-2 px-4 py-2 text-gray-600 hover:text-gray-800 hover:bg-gray-100 rounded-xl transition-all"
            >
              <ArrowLeft className="w-4 h-4" />
              Nazad
            </button>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 grid grid-cols-1 lg:grid-cols-4 gap-8">
        {/* Sidebar */}
        <nav className="lg:col-span-1 bg-white rounded-2xl shadow-lg p-3 h-fit">
          {SECTIONS.map((s) => {
            const Icon = s.icon;
            const isActive = active === s.id;
            return (
              <button
                key={s.id}
                onClick={() => setActive(s.id)}
                className={`w-full flex items-center gap-3 px-4 py-3 rounded-xl text-sm font-medium transition-all mb-1 ${
                  isActive
                    ? "bg-gradient-to-r from-blue-600 to-purple-600 text-white shadow-md"
                    : "text-gray-700 hover:bg-gray-100"
                }`}
              >
                <Icon className="w-4 h-4" />
                {s.label}
              </button>
            );
          })}
        </nav>

        {/* Content */}
        <section className="lg:col-span-3 space-y-6">
          {active === "profile" && (
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-lg font-bold text-gray-800 mb-6 flex items-center gap-2">
                <User className="w-5 h-5 text-blue-600" /> Profil
              </h2>
              <div className="flex items-center gap-4 mb-6">
                <div className="flex items-center justify-center w-16 h-16 bg-gradient-to-br from-blue-600 to-purple-600 rounded-full shadow-md">
                  <span className="text-white text-2xl font-semibold">
                    {userInfo?.email?.charAt(0).toUpperCase() || "U"}
                  </span>
                </div>
                <div>
                  <p className="text-lg font-semibold text-gray-800">{userInfo?.email}</p>
                  <span
                    className={`inline-block mt-1 text-xs font-semibold px-2.5 py-1 rounded-full ${
                      userInfo?.role === "admin"
                        ? "bg-purple-100 text-purple-700"
                        : "bg-blue-100 text-blue-700"
                    }`}
                  >
                    {ROLE_LABELS[userInfo?.role] || userInfo?.role}
                  </span>
                </div>
              </div>
              <dl className="divide-y divide-gray-100 text-sm">
                <div className="flex justify-between py-3">
                  <dt className="text-gray-500">Korisnički ID</dt>
                  <dd className="font-mono text-gray-800">{userInfo?.user_id}</dd>
                </div>
                <div className="flex justify-between py-3">
                  <dt className="text-gray-500">Email</dt>
                  <dd className="text-gray-800">{userInfo?.email}</dd>
                </div>
                <div className="flex justify-between py-3">
                  <dt className="text-gray-500">Rola</dt>
                  <dd className="text-gray-800">{ROLE_LABELS[userInfo?.role] || userInfo?.role}</dd>
                </div>
              </dl>
            </div>
          )}

          {active === "security" && (
            <>
              <div className="bg-white rounded-2xl shadow-lg p-6">
                <h2 className="text-lg font-bold text-gray-800 mb-6 flex items-center gap-2">
                  <Lock className="w-5 h-5 text-blue-600" /> Promena lozinke
                </h2>
                <div className="space-y-4">
                  <Field
                    label="Trenutna lozinka"
                    type="password"
                    value={pwCurrent}
                    onChange={(e) => setPwCurrent(e.target.value)}
                  />
                  <Field
                    label="Nova lozinka"
                    type="password"
                    value={pwNew}
                    onChange={(e) => setPwNew(e.target.value)}
                  />
                  <Field
                    label="Potvrdi novu lozinku"
                    type="password"
                    value={pwConfirm}
                    onChange={(e) => setPwConfirm(e.target.value)}
                  />
                  <button
                    onClick={changePassword}
                    disabled={pwBusy || !pwCurrent || !pwNew || !pwConfirm}
                    className="bg-gradient-to-r from-blue-600 to-purple-600 text-white px-6 py-2.5 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
                  >
                    {pwBusy ? "Čuvanje..." : "Promeni lozinku"}
                  </button>
                  <Toast msg={pwMsg} />
                </div>
              </div>

              <div className="bg-white rounded-2xl shadow-lg p-6">
                <h2 className="text-lg font-bold text-gray-800 mb-6 flex items-center gap-2">
                  <Mail className="w-5 h-5 text-blue-600" /> Promena email adrese
                </h2>
                <div className="space-y-4">
                  <Field
                    label="Nova email adresa"
                    type="email"
                    placeholder="novi@primer.com"
                    value={newEmail}
                    onChange={(e) => setNewEmail(e.target.value)}
                  />
                  <Field
                    label="Trenutna lozinka (potvrda)"
                    type="password"
                    value={emailPw}
                    onChange={(e) => setEmailPw(e.target.value)}
                  />
                  <button
                    onClick={changeEmail}
                    disabled={emailBusy || !newEmail || !emailPw}
                    className="bg-gradient-to-r from-blue-600 to-purple-600 text-white px-6 py-2.5 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
                  >
                    {emailBusy ? "Čuvanje..." : "Promeni email"}
                  </button>
                  <Toast msg={emailMsg} />
                </div>
              </div>
            </>
          )}

          {active === "display" && (
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-lg font-bold text-gray-800 mb-2 flex items-center gap-2">
                <SlidersHorizontal className="w-5 h-5 text-blue-600" /> Pragovi upozorenja
              </h2>
              <p className="text-sm text-gray-500 mb-6">
                Boje na karticama vozila zavise od ovih pragova.
              </p>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-x-6 gap-y-4">
                <Field label="Motor — upozorenje (°C)" type="number" value={draft.engineWarn} onChange={setNum("engineWarn")} />
                <Field label="Motor — kritično (°C)" type="number" value={draft.engineCritical} onChange={setNum("engineCritical")} />
                <Field label="Rashladna — upozorenje (°C)" type="number" value={draft.coolantWarn} onChange={setNum("coolantWarn")} />
                <Field label="Rashladna — kritično (°C)" type="number" value={draft.coolantCritical} onChange={setNum("coolantCritical")} />
                <Field label="Gorivo — upozorenje (%)" type="number" value={draft.fuelWarn} onChange={setNum("fuelWarn")} />
                <Field label="Gorivo — kritično (%)" type="number" value={draft.fuelCritical} onChange={setNum("fuelCritical")} />
              </div>

              <div className="flex items-center gap-3 mt-6">
                <button
                  onClick={savePrefs}
                  className="bg-gradient-to-r from-blue-600 to-purple-600 text-white px-6 py-2.5 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transition-all shadow-lg"
                >
                  Sačuvaj
                </button>
                <button
                  onClick={resetPrefs}
                  className="flex items-center gap-2 px-4 py-2.5 text-gray-600 hover:bg-gray-100 rounded-xl transition-all font-medium"
                >
                  <RotateCcw className="w-4 h-4" />
                  Podrazumevano
                </button>
              </div>
              <div className="mt-4">
                <Toast msg={prefMsg} />
              </div>
            </div>
          )}

          {active === "system" && (
            <div className="bg-white rounded-2xl shadow-lg p-6">
              <h2 className="text-lg font-bold text-gray-800 mb-6 flex items-center gap-2">
                <Server className="w-5 h-5 text-blue-600" /> Sistem i sesija
              </h2>
              <dl className="divide-y divide-gray-100 text-sm">
                <div className="flex justify-between items-center py-3">
                  <dt className="text-gray-500">WebSocket veza</dt>
                  <dd
                    className={`flex items-center gap-2 font-medium ${
                      wsConnected ? "text-green-600" : "text-red-600"
                    }`}
                  >
                    {wsConnected ? <Wifi className="w-4 h-4" /> : <WifiOff className="w-4 h-4" />}
                    {wsConnected ? "Povezano (Live)" : "Nije povezano"}
                  </dd>
                </div>
                <div className="flex justify-between py-3">
                  <dt className="text-gray-500">API adresa</dt>
                  <dd className="font-mono text-gray-800">{API_URL}</dd>
                </div>
                <div className="flex justify-between py-3">
                  <dt className="text-gray-500">Verzija aplikacije</dt>
                  <dd className="text-gray-800">{APP_VERSION}</dd>
                </div>
              </dl>

              <div className="flex flex-col sm:flex-row gap-3 mt-6">
                {onOpenAdmin && (
                  <button
                    onClick={onOpenAdmin}
                    className="flex items-center justify-center gap-2 px-5 py-2.5 bg-blue-50 text-blue-700 rounded-xl font-semibold hover:bg-blue-100 transition-all"
                  >
                    <ShieldCheck className="w-4 h-4" />
                    Otvori administraciju
                  </button>
                )}
                <button
                  onClick={onLogout}
                  className="flex items-center justify-center gap-2 px-5 py-2.5 bg-red-50 text-red-600 rounded-xl font-semibold hover:bg-red-100 transition-all"
                >
                  <LogOut className="w-4 h-4" />
                  Odjavi se
                </button>
              </div>
            </div>
          )}
        </section>
      </main>
    </div>
  );
}
