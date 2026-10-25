import type {Metadata} from "next";

// The page is a Client Component, which can't export metadata, so the title lives here.
// Detail pages below override it with their own generateMetadata.
export const metadata: Metadata = {
    title: "Products",
};

export default function ProductsLayout({children}: LayoutProps<"/products">) {
    return children;
}
