import type {Metadata} from "next";

// The page is a Client Component, which can't export metadata, so the title lives here.
// Detail pages below override it with their own generateMetadata.
export const metadata: Metadata = {
    title: "Orders",
};

export default function OrdersLayout({children}: LayoutProps<"/orders">) {
    return children;
}
