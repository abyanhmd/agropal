import React from 'react';
import DashboardLoader from "@/app/(app)/dashboard/_components/dashboard-loader";
import type {Metadata} from "next";

export const metadata: Metadata = {
    title: "Dashboard",
};

export default function DashboardPage() {
    return <DashboardLoader />;
}
