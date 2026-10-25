import React from 'react';
import type {Metadata} from "next";
import ProductLoader from "@/app/(app)/products/_components/loader.product";

// Products come from the signed-in user's farm, which only the browser can load
export const metadata: Metadata = {
    title: "Product",
};

export default async function ProductDetailPage({params}: PageProps<"/products/[id]">) {
    const {id} = await params;
    return <ProductLoader id={id} />;
}
