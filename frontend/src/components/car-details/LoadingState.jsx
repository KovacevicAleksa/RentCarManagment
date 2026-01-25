import { Activity } from "lucide-react";

export default function LoadingState({ carId }) {
  return (
    <div className="bg-white rounded-3xl shadow-xl p-12 text-center">
      <Activity className="w-16 h-16 text-gray-400 mx-auto mb-4 animate-pulse" />
      <h3 className="text-xl font-bold text-gray-800 mb-2">
        Učitavanje podataka...
      </h3>
      <p className="text-gray-600">Čekanje telemetrije vozila {carId}</p>
    </div>
  );
}
