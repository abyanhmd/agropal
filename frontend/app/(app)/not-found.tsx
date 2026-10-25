import NotFoundContent from "@/app/_components/not-found-content";
import NotFoundBreadcrumb from "@/app/(app)/_components/not-found-breadcrumb";

// notFound() from an app route, e.g. /products/unknown-id: keeps the sidebar.
// Those pages set the tab title themselves via generateMetadata.
export default function AppNotFound() {
    return (
        <>
            <NotFoundBreadcrumb />
            <NotFoundContent homeHref="/dashboard" />
        </>
    );
}
