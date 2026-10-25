import {type ReactNode, Suspense} from "react";
import type {Metadata} from "next";
import {SidebarInset, SidebarProvider, SidebarTrigger} from "@/components/ui/sidebar";
import AppSidebar from "@/components/app/app-sidebar";
import AppBreadcrumb from "@/components/app/app-breadcrumb";
import {Separator} from "@/components/ui/separator";
import {BreadcrumbProvider} from "@/hooks/breadcrumb-state";
import {DemoAuthProvider} from "@/lib/auth/auth-provider";
import {AppBaseProvider, DEMO_BASE} from "@/lib/demo/base";
import DemoBanner from "@/app/(demo)/demo/_components/demo-banner";

export const metadata: Metadata = {
    title: {default: "Demo", template: "%s · Demo"},
};

// The same app shell as (app), signed in as a made-up farmer and fed sample data. No account needed:
// requests made under /demo are answered from lib/demo instead of the backend.
export default function DemoLayout({children}: {children: ReactNode}) {
    return (
        <AppBaseProvider base={DEMO_BASE}>
        <DemoAuthProvider>
        <SidebarProvider>
            <BreadcrumbProvider>
                <AppSidebar />
                <SidebarInset className="min-w-0">
                    <DemoBanner />
                    <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator orientation="vertical" className="mr-2 h-4 self-center" />
                        <Suspense fallback={<div className="h-5" />}>
                            <AppBreadcrumb />
                        </Suspense>
                    </header>
                    {/* Detail pages read their params, which are only known per request */}
                    <Suspense fallback={<div className="page-container"><div className="h-64 animate-pulse rounded-xl bg-muted" /></div>}>
                        {children}
                    </Suspense>
                </SidebarInset>
            </BreadcrumbProvider>
        </SidebarProvider>
        </DemoAuthProvider>
        </AppBaseProvider>
    );
}
