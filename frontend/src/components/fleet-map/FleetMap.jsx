import { useEffect, useRef, useState } from "react";
import { MapPin, X, Car, Thermometer, Droplet, Gauge } from "lucide-react";

export default function FleetMap({ cars, onClose }) {
  const mapContainerRef = useRef(null);
  const mapInstanceRef = useRef(null);
  const markersRef = useRef({});
  const [leafletLoaded, setLeafletLoaded] = useState(false);
  const [selectedCarId, setSelectedCarId] = useState(null);

  // Load Leaflet library
  useEffect(() => {
    if (window.L) {
      setLeafletLoaded(true);
      return;
    }

    const link = document.createElement("link");
    link.rel = "stylesheet";
    link.href =
      "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/leaflet.css";
    document.head.appendChild(link);

    const script = document.createElement("script");
    script.src =
      "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/leaflet.js";
    script.async = true;

    script.onload = () => {
      setLeafletLoaded(true);
    };

    script.onerror = () => {
      console.error("❌ Greška pri učitavanju Leaflet biblioteke");
    };

    document.body.appendChild(script);

    return () => {};
  }, []);

  // Initialize map
  useEffect(() => {
    if (!leafletLoaded || !window.L || !mapContainerRef.current) {
      return;
    }

    if (mapInstanceRef.current) {
      mapInstanceRef.current.remove();
      mapInstanceRef.current = null;
      markersRef.current = {};
    }

    const timer = setTimeout(() => {
      try {
        const mapInstance = window.L.map(mapContainerRef.current, {
          center: [44.813346, 20.409719],
          zoom: 12,
          zoomControl: true,
          attributionControl: true,
        });

        mapInstanceRef.current = mapInstance;

        window.L.tileLayer(
          "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
          {
            attribution:
              '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            maxZoom: 19,
          },
        ).addTo(mapInstance);

        setTimeout(() => {
          if (mapInstance) {
            mapInstance.invalidateSize();
          }
        }, 300);
      } catch (error) {
        console.error("❌ Greška pri inicijalizaciji mape:", error);
      }
    }, 150);

    return () => {
      clearTimeout(timer);
      if (mapInstanceRef.current) {
        mapInstanceRef.current.remove();
        mapInstanceRef.current = null;
        markersRef.current = {};
      }
    };
  }, [leafletLoaded]);

  // Update markers for all cars
  useEffect(() => {
    if (!mapInstanceRef.current || !window.L || !cars) return;

    try {
      const carArray = Object.values(cars);
      const currentCarIds = Object.keys(cars);

      // Remove markers for cars that no longer exist
      Object.keys(markersRef.current).forEach((carId) => {
        if (!currentCarIds.includes(carId)) {
          markersRef.current[carId].remove();
          delete markersRef.current[carId];
        }
      });

      // Add/update markers for each car
      carArray.forEach((car) => {
        const { car_id, latitude, longitude, engine_temperature, fuel_level } =
          car;

        // Determine marker color based on status
        let markerColor = "#3b82f6"; // blue default
        if (engine_temperature >= 100) {
          markerColor = "#dc2626"; // red
        } else if (engine_temperature >= 90) {
          markerColor = "#f59e0b"; // yellow
        } else if (fuel_level <= 15) {
          markerColor = "#dc2626"; // red
        } else if (fuel_level <= 30) {
          markerColor = "#f59e0b"; // yellow
        }

        const carIcon = window.L.divIcon({
          html: `<div style="background: ${markerColor}; width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; border: 3px solid white; box-shadow: 0 4px 12px rgba(0,0,0,0.3);">
            <svg xmlns="http://www.w3.org/2000/svg" width="26" height="26" viewBox="0 0 24 24" fill="white" stroke="none">
              <path d="M5 11l1.5-4.5h11L19 11m-1.5 5a1.5 1.5 0 0 1-1.5-1.5a1.5 1.5 0 0 1 1.5-1.5a1.5 1.5 0 0 1 1.5 1.5a1.5 1.5 0 0 1-1.5 1.5m-11 0A1.5 1.5 0 0 1 5 14.5A1.5 1.5 0 0 1 6.5 13A1.5 1.5 0 0 1 8 14.5A1.5 1.5 0 0 1 6.5 16M18.92 6c-.2-.58-.76-1-1.42-1h-11c-.66 0-1.22.42-1.42 1L3 12v8a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-1h12v1a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-8l-2.08-6z"/>
            </svg>
          </div>`,
          className: "",
          iconSize: [44, 44],
          iconAnchor: [22, 22],
        });

        if (markersRef.current[car_id]) {
          // Update existing marker - ažuriraj poziciju, ikonu i popup
          markersRef.current[car_id].setLatLng([latitude, longitude]);
          markersRef.current[car_id].setIcon(carIcon);
          markersRef.current[car_id].setPopupContent(createPopupContent(car));
        } else {
          // Create new marker
          const newMarker = window.L.marker([latitude, longitude], {
            icon: carIcon,
          }).addTo(mapInstanceRef.current);

          newMarker.bindPopup(createPopupContent(car), { maxWidth: 250 });

          newMarker.on("click", () => {
            setSelectedCarId(car_id);
          });

          markersRef.current[car_id] = newMarker;
        }
      });

      // Fit bounds to show all markers
      if (carArray.length > 0) {
        const bounds = window.L.latLngBounds(
          carArray.map((car) => [car.latitude, car.longitude]),
        );
        mapInstanceRef.current.fitBounds(bounds, { padding: [50, 50] });
      }
    } catch (error) {
      console.error("❌ Greška pri ažuriranju markera:", error);
    }
  }, [cars]);

  const createPopupContent = (car) => {
    return `
      <div style="text-align: left; min-width: 200px;">
        <h3 style="font-weight: bold; font-size: 16px; margin-bottom: 8px; color: #1f2937;">${car.car_id}</h3>
        <div style="font-size: 12px; color: #6b7280; margin-bottom: 8px;">
          <div>📍 ${car.latitude.toFixed(6)}°, ${car.longitude.toFixed(6)}°</div>
        </div>
        <div style="display: flex; flex-direction: column; gap: 4px; font-size: 13px;">
          <div style="display: flex; justify-content: space-between;">
            <span>🌡️ Motor:</span>
            <strong style="color: ${car.engine_temperature >= 100 ? "#dc2626" : car.engine_temperature >= 90 ? "#f59e0b" : "#10b981"}">${car.engine_temperature.toFixed(1)}°C</strong>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span>💧 Rashladna:</span>
            <strong style="color: ${car.engine_coolant_temp >= 95 ? "#dc2626" : car.engine_coolant_temp >= 85 ? "#f59e0b" : "#10b981"}">${car.engine_coolant_temp.toFixed(1)}°C</strong>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span>⛽ Gorivo:</span>
            <strong style="color: ${car.fuel_level <= 15 ? "#dc2626" : car.fuel_level <= 30 ? "#f59e0b" : "#10b981"}">${car.fuel_level.toFixed(1)}%</strong>
          </div>
          <div style="display: flex; justify-content: space-between;">
            <span>🚗 Gas:</span>
            <strong style="color: #3b82f6">${(car.gas_throttle * 100).toFixed(0)}%</strong>
          </div>
        </div>
      </div>
    `;
  };

  const selectedCar = selectedCarId ? cars[selectedCarId] : null;

  // Calculate car counts by status
  const getCarStatusCounts = () => {
    const counts = { critical: 0, warning: 0, normal: 0 };
    Object.values(cars).forEach((car) => {
      if (car.engine_temperature >= 100 || car.fuel_level <= 15) {
        counts.critical++;
      } else if (car.engine_temperature >= 90 || car.fuel_level <= 30) {
        counts.warning++;
      } else {
        counts.normal++;
      }
    });
    return counts;
  };

  const statusCounts = getCarStatusCounts();

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center p-4">
      <div className="bg-white rounded-2xl shadow-2xl w-full h-full max-w-7xl max-h-[90vh] flex flex-col">
        {/* Header */}
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 text-white p-6 rounded-t-2xl flex items-center justify-between">
          <div className="flex items-center gap-3">
            <MapPin className="w-8 h-8" />
            <div>
              <h2 className="text-2xl font-bold">Mapa Flote</h2>
              <p className="text-blue-100 text-sm">
                {Object.keys(cars).length} aktivnih vozila
              </p>
            </div>
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-white hover:bg-opacity-20 rounded-lg transition-all"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        {/* Map Container */}
        <div className="flex-1 relative">
          <div
            ref={mapContainerRef}
            className="w-full h-full"
            style={{ minHeight: "400px" }}
          />

          {!leafletLoaded && (
            <div className="absolute inset-0 flex items-center justify-center bg-gray-100">
              <div className="text-center">
                <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-2"></div>
                <p className="text-gray-600">Učitavanje mape...</p>
              </div>
            </div>
          )}

          {/* Selected Car Info Panel */}
          {selectedCar && (
            <div className="absolute top-4 right-4 bg-white rounded-xl shadow-xl p-4 max-w-sm z-10">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <Car className="w-5 h-5 text-blue-600" />
                  <h3 className="font-bold text-lg">{selectedCar.car_id}</h3>
                </div>
                <button
                  onClick={() => setSelectedCarId(null)}
                  className="p-1 hover:bg-gray-100 rounded"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <div className="space-y-2 text-sm">
                <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div className="flex items-center gap-2">
                    <Thermometer className="w-4 h-4 text-orange-600" />
                    <span>Motor</span>
                  </div>
                  <span className="font-bold">
                    {selectedCar.engine_temperature.toFixed(1)}°C
                  </span>
                </div>

                <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div className="flex items-center gap-2">
                    <Droplet className="w-4 h-4 text-blue-600" />
                    <span>Rashladna</span>
                  </div>
                  <span className="font-bold">
                    {selectedCar.engine_coolant_temp.toFixed(1)}°C
                  </span>
                </div>

                <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div className="flex items-center gap-2">
                    <Droplet className="w-4 h-4 text-green-600" />
                    <span>Gorivo</span>
                  </div>
                  <span className="font-bold">
                    {selectedCar.fuel_level.toFixed(1)}%
                  </span>
                </div>

                <div className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div className="flex items-center gap-2">
                    <Gauge className="w-4 h-4 text-purple-600" />
                    <span>Gas</span>
                  </div>
                  <span className="font-bold">
                    {(selectedCar.gas_throttle * 100).toFixed(0)}%
                  </span>
                </div>

                <div className="pt-2 border-t">
                  <div className="text-xs text-gray-600">
                    📍 {selectedCar.latitude.toFixed(6)}°,{" "}
                    {selectedCar.longitude.toFixed(6)}°
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Footer with car count */}
        <div className="bg-gray-50 p-4 rounded-b-2xl border-t">
          <div className="flex items-center justify-between text-sm text-gray-600">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-green-500"></div>
              <span>Normalno ({statusCounts.normal})</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-yellow-500"></div>
              <span>Upozorenje ({statusCounts.warning})</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-red-500"></div>
              <span>Kritično ({statusCounts.critical})</span>
            </div>
            <span className="font-semibold">
              Ukupno: {Object.keys(cars).length} vozila
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}
