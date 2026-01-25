import { useEffect, useRef, useState } from "react";
import { MapPin } from "lucide-react";

export default function CarMap({ carId, carData }) {
  const mapContainerRef = useRef(null);
  const mapInstanceRef = useRef(null);
  const markerRef = useRef(null);
  const [leafletLoaded, setLeafletLoaded] = useState(false);

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
      console.log("✅ Leaflet učitan");
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
      console.log("⏳ Čekanje na Leaflet...");
      return;
    }

    // Cleanup existing map
    if (mapInstanceRef.current) {
      mapInstanceRef.current.remove();
      mapInstanceRef.current = null;
      markerRef.current = null;
    }

    const timer = setTimeout(() => {
      try {
        console.log("🗺️ Kreiranje mape...");

        const mapInstance = window.L.map(mapContainerRef.current, {
          center: [44.813346, 20.409719],
          zoom: 13,
          zoomControl: true,
          attributionControl: true,
        });

        mapInstanceRef.current = mapInstance;

        // Add tile layer (OpenStreetMap)
        window.L.tileLayer(
          "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
          {
            attribution:
              '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
            maxZoom: 19,
          },
        ).addTo(mapInstance);

        // Critical: invalidate size after map is added to DOM
        setTimeout(() => {
          if (mapInstance) {
            mapInstance.invalidateSize();
            console.log("✅ Mapa uspešno inicijalizovana za", carId);
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
        markerRef.current = null;
      }
    };
  }, [carId, leafletLoaded]);

  // Update map location
  useEffect(() => {
    if (!carData || !mapInstanceRef.current || !window.L) return;

    try {
      const { latitude, longitude } = carData;

      mapInstanceRef.current.setView([latitude, longitude], 15, {
        animate: true,
        duration: 1,
      });

      if (markerRef.current) {
        markerRef.current.setLatLng([latitude, longitude]);
        markerRef.current.setPopupContent(
          `<div style="text-align: center;">
            <strong>${carId}</strong><br/>
            Širina: ${latitude.toFixed(6)}°<br/>
            Dužina: ${longitude.toFixed(6)}°
          </div>`,
        );
      } else {
        const carIcon = window.L.divIcon({
          html: `<div style="background: #3b82f6; width: 44px; height: 44px; border-radius: 8px; display: flex; align-items: center; justify-content: center; border: 3px solid white; box-shadow: 0 4px 12px rgba(0,0,0,0.3);">
            <svg xmlns="http://www.w3.org/2000/svg" width="26" height="26" viewBox="0 0 24 24" fill="white" stroke="none">
              <path d="M5 11l1.5-4.5h11L19 11m-1.5 5a1.5 1.5 0 0 1-1.5-1.5a1.5 1.5 0 0 1 1.5-1.5a1.5 1.5 0 0 1 1.5 1.5a1.5 1.5 0 0 1-1.5 1.5m-11 0A1.5 1.5 0 0 1 5 14.5A1.5 1.5 0 0 1 6.5 13A1.5 1.5 0 0 1 8 14.5A1.5 1.5 0 0 1 6.5 16M18.92 6c-.2-.58-.76-1-1.42-1h-11c-.66 0-1.22.42-1.42 1L3 12v8a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-1h12v1a1 1 0 0 0 1 1h1a1 1 0 0 0 1-1v-8l-2.08-6z"/>
            </svg>
          </div>`,
          className: "",
          iconSize: [44, 44],
          iconAnchor: [22, 22],
        });

        const newMarker = window.L.marker([latitude, longitude], {
          icon: carIcon,
        }).addTo(mapInstanceRef.current);

        newMarker.bindPopup(
          `<div style="text-align: center;">
            <strong>${carId}</strong><br/>
            Širina: ${latitude.toFixed(6)}°<br/>
            Dužina: ${longitude.toFixed(6)}°
          </div>`,
          { maxWidth: 200 },
        );

        markerRef.current = newMarker;
      }
    } catch (error) {
      console.error("❌ Greška pri ažuriranju mape:", error);
    }
  }, [carData, carId]);

  if (!carData) return null;

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-800 mb-2 flex items-center gap-2">
          <MapPin className="w-5 h-5 text-blue-500" />
          Lokacija u realnom vremenu
        </h3>
        <div className="text-sm text-gray-600 space-y-1">
          <div>Širina: {carData.latitude.toFixed(6)}°</div>
          <div>Dužina: {carData.longitude.toFixed(6)}°</div>
        </div>
      </div>

      <div
        ref={mapContainerRef}
        className="w-full rounded-lg overflow-hidden border border-gray-200"
        style={{ height: "400px" }}
      />

      {!leafletLoaded && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 rounded-lg">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500 mx-auto mb-2"></div>
            <p className="text-gray-600">Učitavanje mape...</p>
          </div>
        </div>
      )}
    </div>
  );
}
