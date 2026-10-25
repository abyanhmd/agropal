import type {Metadata} from "next";
import CalendarPage from "@/app/(app)/calendar/page";

export const metadata: Metadata = {title: "Crop calendar"};

export default function DemoCalendarPage() {
    return <CalendarPage />;
}
