import React from 'react';
import type {Metadata} from "next";
import OrderLoader from "@/app/(app)/orders/_components/loader.order";

// Orders come from the signed-in user's farm, which only the browser can load
export const metadata: Metadata = {
    title: "Order",
};

export default async function OrderDetailPage({params}: PageProps<"/orders/[id]">) {
    const {id} = await params;
    return <OrderLoader id={id} />;
}
