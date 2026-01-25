import { useWebSocket } from "../../contexts/WebSocketContext";
import CarDetailsHeader from "./CarDetailsHeader";
import CarMap from "./CarMap";
import CarInfoPanel from "./CarInfoPanel";
import LoadingState from "./LoadingState";
import CarHistoryCharts from "./CarHistoryCharts";

export default function CarDetails({ carId, onBack }) {
  const { getCar, wsConnected } = useWebSocket();
  const carData = getCar(carId);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-purple-50">
      <CarDetailsHeader
        carId={carId}
        wsConnected={wsConnected}
        onBack={onBack}
      />

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {!carData ? (
          <LoadingState carId={carId} />
        ) : (
          <>
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              <div className="lg:col-span-2">
                <CarMap carId={carId} carData={carData} />
              </div>
              <CarInfoPanel carData={carData} />
            </div>
            <CarHistoryCharts carId={carId} />
          </>
        )}
      </main>
    </div>
  );
}
