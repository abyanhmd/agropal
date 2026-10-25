import type {Metadata} from "next";
import DashboardLoader from "@/app/(app)/dashboard/_components/dashboard-loader";

export const metadata: Metadata = {title: "Dashboard"};

export default function DemoDashboardPage() {
    return <DashboardLoader />;
}
