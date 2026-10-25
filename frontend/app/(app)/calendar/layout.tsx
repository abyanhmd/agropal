import type {Metadata} from "next";

// The page is a Client Component, which can't export metadata, so the title lives here.
export const metadata: Metadata = {
    title: "Crop calendar",
};

export default function CalendarLayout({children}: LayoutProps<"/calendar">) {
    return children;
}
