import {type ReactNode, Suspense} from "react";
import {SidebarInset, SidebarProvider, SidebarTrigger} from "@/components/ui/sidebar";
import AppSidebar from "@/components/app/app-sidebar";
import AppBreadcrumb from "@/components/app/app-breadcrumb";
import {Separator} from "@/components/ui/separator";
import {BreadcrumbProvider} from "@/hooks/breadcrumb-state";
import {AuthProvider} from "@/lib/auth/auth-provider";
import AuthGate from "@/lib/auth/auth-gate";

// The app shell (sidebar + breadcrumb header). The landing page at "/" sits outside this group.
export default function AppLayout({children}: {children: ReactNode}) {
    return (
        <AuthProvider>
        <AuthGate>
        <SidebarProvider>
            <BreadcrumbProvider>
                <AppSidebar />
                <SidebarInset className="min-w-0">
                    <header className="flex h-14 shrink-0 items-center gap-2 border-b border-border px-4">
                        <SidebarTrigger className="-ml-1" />
                        <Separator orientation="vertical" className="mr-2 h-4 self-center" />
                        <Suspense fallback={<div className="h-5" />}>
                            <AppBreadcrumb />
                        </Suspense>
                    </header>
                    {children}
                </SidebarInset>
            </BreadcrumbProvider>
        </SidebarProvider>
        </AuthGate>
        </AuthProvider>
    );
}
