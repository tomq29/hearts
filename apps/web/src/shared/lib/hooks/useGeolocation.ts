import { useState } from "react";

interface Coords {
  latitude: number;
  longitude: number;
}

/**
 * Reads the browser's location on demand. Requires a secure context, so it
 * works on localhost in dev but needs HTTPS once deployed.
 */
export function useGeolocation() {
  const [isLocating, setIsLocating] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const request = (onSuccess: (coords: Coords) => void) => {
    if (!navigator.geolocation) {
      setError("Your browser does not support location sharing.");
      return;
    }

    setIsLocating(true);
    setError(null);

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setIsLocating(false);
        onSuccess({
          latitude: position.coords.latitude,
          longitude: position.coords.longitude,
        });
      },
      (err) => {
        setIsLocating(false);
        setError(
          err.code === err.PERMISSION_DENIED
            ? "Location permission denied. You can enable it in your browser settings."
            : "Could not determine your location. Please try again."
        );
      },
      { timeout: 10000 }
    );
  };

  return { request, isLocating, error };
}
