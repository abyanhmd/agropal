'use client';

import {useEffect, useState} from 'react';
import {ForecastSchema} from "@/app/(app)/weather/_types/weather.schema";
import {getForecast} from "@/app/(app)/weather/_lib/open-meteo";

export type ForecastState =
    | {status: "loading"}
    | {status: "ready", forecast: ForecastSchema}
    | {status: "error", message: string};

type KeyedState = ForecastState & {key: string};

// Open-Meteo forecast for a point, shared by the weather page and the dashboard card
export function useForecast(latitude: number, longitude: number): ForecastState & {retry: () => void} {
    const key = `${latitude},${longitude}`;
    const [state, setState] = useState<KeyedState>({status: "loading", key});
    const [attempt, setAttempt] = useState(0);

    useEffect(() => {
        const controller = new AbortController();
        getForecast(latitude, longitude, controller.signal)
            .then((forecast) => setState({status: "ready", key, forecast}))
            .catch((error: unknown) => {
                if (controller.signal.aborted) return;
                setState({status: "error", key, message: error instanceof Error ? error.message : "Something went wrong"});
            });
        return () => controller.abort();
    }, [latitude, longitude, key, attempt]);

    const retry = () => {
        setState({status: "loading", key});
        setAttempt((n) => n + 1);
    };

    // A moved pin reads as loading until its own forecast arrives, never the old location's
    return state.key === key ? {...state, retry} : {status: "loading", retry};
}

// Weekly rain minus reference evapotranspiration (FAO-56 ET₀): negative means a water deficit
export function weeklyWaterBalance(forecast: ForecastSchema) {
    const rain = forecast.daily.precipitation_sum.reduce((total, v) => total + v, 0);
    const et0 = forecast.daily.et0_fao_evapotranspiration.reduce((total, v) => total + v, 0);
    return {rain, et0, balance: rain - et0};
}
