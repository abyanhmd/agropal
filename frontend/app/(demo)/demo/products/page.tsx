import type {Metadata} from "next";
import ProductPage from "@/app/(app)/products/page";

export const metadata: Metadata = {title: "Products"};

export default function DemoProductsPage() {
    return <ProductPage />;
}
