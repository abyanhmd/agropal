import type {Metadata} from "next";
import OrderPage from "@/app/(app)/orders/page";

export const metadata: Metadata = {title: "Orders"};

export default function DemoOrdersPage() {
    return <OrderPage />;
}
