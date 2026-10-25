import type {Metadata} from "next";
import ProductLoader from "@/app/(app)/products/_components/loader.product";

export const metadata: Metadata = {title: "Product"};

export default async function DemoProductPage({params}: {params: Promise<{id: string}>}) {
    const {id} = await params;
    return <ProductLoader id={id} />;
}
