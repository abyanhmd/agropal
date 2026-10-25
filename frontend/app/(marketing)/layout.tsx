import type {ReactNode} from "react";
import Header from "@/app/(marketing)/_components/header.landing";
import Footer from "@/app/(marketing)/_components/footer.landing";

// Public pages (the landing page, and later e.g. /about or /pricing) share this header and footer.
// The app itself lives in the (app) group with its own sidebar layout.
export default function MarketingLayout({children}: {children: ReactNode}) {
    return (
        <>
            <Header />
            <main className="flex flex-1 flex-col overflow-x-clip">{children}</main>
            <Footer />
        </>
    );
}
