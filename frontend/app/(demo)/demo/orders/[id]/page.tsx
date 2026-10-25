import type {Metadata} from "next";
import OrderLoader from "@/app/(app)/orders/_components/loader.order";

export const metadata: Metadata = {title: "Order"};

export default async function DemoOrderPage({params}: {params: Promise<{id: string}>}) {
    const {id} = await params;
    return <OrderLoader id={id} />;
}
