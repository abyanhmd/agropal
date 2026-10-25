'use client';

import React, {useState} from 'react';
import {useFarmLocation} from "@/app/(app)/weather/_lib/location-store";
import LocationPicker, {LocationPickerSkeleton} from "@/app/(app)/weather/_components/location-picker.weather";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

// The land location behind the weather and crop calendar pages
export default function LocationSettings() {
    const {location, save, ready} = useFarmLocation();
    const [error, setError] = useState<string | null>(null);

    if (!ready) return <LocationPickerSkeleton />;

    return (
        <>
            <LocationPicker
                saved={location}
                onSave={(latitude, longitude, name) => {
                    setError(null);
                    save(latitude, longitude, name).catch((err: unknown) => setError(errorMessage(err)));
                }}
            />
            {error && <p role="alert" className="text-sm text-danger-strong font-sans">Couldn&apos;t save the location: {error}</p>}
        </>
    );
}
