import type {Metadata} from "next";

// The page is a Client Component, which can't export metadata, so the title lives here.
export const metadata: Metadata = {
    title: "Weather",
};

export default function WeatherLayout({children}: LayoutProps<"/weather">) {
    return children;
}
