import type {Metadata} from "next";
import WeatherPage from "@/app/(app)/weather/page";

export const metadata: Metadata = {title: "Weather"};

export default function DemoWeatherPage() {
    return <WeatherPage />;
}
