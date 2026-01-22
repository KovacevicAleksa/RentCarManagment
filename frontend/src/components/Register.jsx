import { useState } from "react";
import { Lock, Mail, Eye, EyeOff } from "lucide-react";

const API_URL = import.meta.env.VITE_API_URL || "http://localhost:8010";
console.log(import.meta.env.VITE_API_URL);

export default function Register() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [showPassword, setShowPassword] = useState(false);
  const [message, setMessage] = useState("");
  const [loading, setLoading] = useState(false);
  const [agreedToTerms, setAgreedToTerms] = useState(false);

  const handleRegister = async () => {
    setMessage("");

    if (!agreedToTerms) {
      setMessage("Morate se složiti sa uslovima korišćenja");
      return;
    }

    if (password !== confirmPassword) {
      setMessage("Lozinke se ne poklapaju");
      return;
    }

    if (password.length < 6) {
      setMessage("Lozinka mora imati najmanje 6 karaktera");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch(`${API_URL}/auth/register`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email, password }),
      });

      if (!res.ok) {
        const err = await res.json();
        setMessage(err.message || "Greška pri registraciji");
        setLoading(false);
        return;
      }

      setMessage("Uspešno ste se registrovali! Sada se možete prijaviti.");
      setEmail("");
      setPassword("");
      setConfirmPassword("");
      setAgreedToTerms(false);
      setLoading(false);
    } catch (err) {
      setMessage("Greška u mreži");
      setLoading(false);
    }
  };

  return (
    <div className="space-y-5">
      {/* Email Input */}
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
            onKeyDown={(e) => e.key === "Enter" && handleRegister()}
            className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
          />
        </div>
      </div>

      {/* Password Input */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Lozinka
        </label>
        <div className="relative">
          <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type={showPassword ? "text" : "password"}
            placeholder="••••••••"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleRegister()}
            className="w-full pl-11 pr-11 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
          />
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
          >
            {showPassword ? (
              <EyeOff className="w-5 h-5" />
            ) : (
              <Eye className="w-5 h-5" />
            )}
          </button>
        </div>
        <p className="text-xs text-gray-500 mt-1">Najmanje 6 karaktera</p>
      </div>

      {/* Confirm Password Input */}
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Potvrdite lozinku
        </label>
        <div className="relative">
          <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-5 h-5 text-gray-400" />
          <input
            type={showPassword ? "text" : "password"}
            placeholder="••••••••"
            value={confirmPassword}
            onChange={(e) => setConfirmPassword(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleRegister()}
            className="w-full pl-11 pr-4 py-3 border border-gray-300 rounded-xl focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none transition-all"
          />
        </div>
      </div>

      {/* Terms */}
      <label className="flex items-start text-sm">
        <input
          type="checkbox"
          checked={agreedToTerms}
          onChange={(e) => setAgreedToTerms(e.target.checked)}
          className="mr-2 mt-1 rounded"
        />
        <span className="text-gray-600">
          Slažem se sa{" "}
          <a href="#" className="text-blue-600 hover:text-blue-700 font-medium">
            uslovima korišćenja
          </a>{" "}
          i{" "}
          <a href="#" className="text-blue-600 hover:text-blue-700 font-medium">
            politikom privatnosti
          </a>
        </span>
      </label>

      {/* Submit Button */}
      <button
        onClick={handleRegister}
        disabled={loading || !email || !password || !confirmPassword}
        className="w-full bg-gradient-to-r from-blue-600 to-purple-600 text-white py-3 rounded-xl font-semibold hover:from-blue-700 hover:to-purple-700 transform hover:scale-[1.02] transition-all disabled:opacity-50 disabled:cursor-not-allowed shadow-lg"
      >
        {loading ? "Registracija u toku..." : "Registruj se"}
      </button>

      {/* Message */}
      {message && (
        <div
          className={`p-4 rounded-xl text-sm ${
            message.includes("Uspešno")
              ? "bg-green-50 text-green-700 border border-green-200"
              : "bg-red-50 text-red-700 border border-red-200"
          }`}
        >
          {message}
        </div>
      )}
    </div>
  );
}
