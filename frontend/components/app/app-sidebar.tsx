'use client';

import React, {Suspense} from 'react';
import {
    CalendarDays,
    ChevronRight,
    CloudSun,
    LayoutDashboard,
    LifeBuoy,
    MessagesSquare,
    Package,
    Settings,
    ShoppingCart,
} from "lucide-react";
import {usePathname} from "next/navigation";
import {
    Sidebar,
    SidebarContent,
    SidebarGroup, SidebarGroupContent, SidebarGroupLabel,
    SidebarFooter,
    SidebarHeader, SidebarMenu,
    SidebarMenuBadge,
    SidebarMenuButton,
    SidebarMenuItem,
    SidebarRail,
    SidebarSeparator,
} from "@/components/ui/sidebar";
import Link from "next/link";
import {useAuth} from "@/lib/auth/auth-provider";
import {useCalendarAvailability} from "@/app/(app)/calendar/_lib/use-calendar";
import {useAppBase} from "@/lib/demo/base";

interface NavItem {
    title: string;
    href: string;
    icon: React.ComponentType;
    badge?: string;
    // Needs crop calendar data for the farm's location; disabled without it
    requiresCalendar?: boolean;
}

interface NavSection {
    label: string;
    items: NavItem[];
}

const sections: NavSection[] = [
    {
        label: "Overview",
        items: [
            {title: "Dashboard", href: "/dashboard", icon: LayoutDashboard},
            {title: "Weather", href: "/weather", icon: CloudSun},
            {title: "Crop calendar", href: "/calendar", icon: CalendarDays, requiresCalendar: true},
        ],
    },
    {
        label: "Commerce",
        items: [
            {title: "Products", href: "/products", icon: Package},
            {title: "Orders", href: "/orders", icon: ShoppingCart},
        ],
    },
    {
        label: "Community",
        items: [
            {title: "Discussions", href: "/discussions", icon: MessagesSquare},
        ],
    },
];

const secondaryNav: NavItem[] = [
    {title: "Settings", href: '/settings', icon: Settings},
];

// Accent bar on the left edge of the active item
const activeBar =
    "relative data-active:before:absolute data-active:before:inset-y-2 data-active:before:left-0 font-medium " +
    "data-active:before:w-0.5 data-active:before:rounded-full data-active:before:bg-sidebar-primary";

// Pure menu; pathname is null while it's unknown (prerendering a dynamic route), so nothing is active yet
function NavItems({items, pathname, calendarAvailable}: {items: NavItem[], pathname: string | null, calendarAvailable: boolean}) {
    // "" in the app, "/demo" in the public demo
    const base = useAppBase();
    const isActive = (href: string) =>
        pathname !== null && pathname.startsWith(base + href);

    return (
        <SidebarMenu>
            {items.map((item) => (
                <SidebarMenuItem key={item.href}>
                    {item.requiresCalendar && !calendarAvailable ? (
                        <SidebarMenuButton
                            tooltip={`${item.title} needs a location with crop calendar data (US only). Set one in Settings.`}
                            aria-disabled
                            className="opacity-50"
                        >
                            <item.icon />
                            <span className="truncate">{item.title}</span>
                        </SidebarMenuButton>
                    ) : (
                        <SidebarMenuButton
                            isActive={isActive(item.href)}
                            tooltip={item.title}
                            className={activeBar}
                            render={<Link href={base + item.href}
                        />}>
                            <item.icon />
                            <span className="truncate">{item.title}</span>
                        </SidebarMenuButton>
                    )}
                    {item.badge && (
                        <SidebarMenuBadge className="rounded-full bg-sidebar-accent text-sidebar-accent-foreground">
                            {item.badge}
                        </SidebarMenuBadge>
                    )}
                </SidebarMenuItem>
            ))}
        </SidebarMenu>
    );
}

// Only this leaf reads the pathname, so only it has to wait for it
function ActiveNavItems({items, calendarAvailable}: {items: NavItem[], calendarAvailable: boolean}) {
    return <NavItems items={items} pathname={usePathname()} calendarAvailable={calendarAvailable} />;
}

function NavMenu({items}: {items: NavItem[]}) {
    const calendarAvailable = useCalendarAvailability() === "available";

    return (
        <Suspense fallback={<NavItems items={items} pathname={null} calendarAvailable={calendarAvailable} />}>
            <ActiveNavItems items={items} calendarAvailable={calendarAvailable} />
        </Suspense>
    );
}

// The signed-in user and their farm; links to settings until there's an account menu
function UserCard() {
    const {session} = useAuth();
    const base = useAppBase();
    if (session.status !== "ready") return null;

    const name = session.user.name || session.user.username;
    const initials = name.split(/\s+/).map((part) => part[0]).slice(0, 2).join("").toUpperCase();

    return (
        <SidebarMenu>
            <SidebarMenuItem>
                <SidebarMenuButton
                    size="lg"
                    tooltip={name}
                    render={<Link href={base + "/settings"} />}
                >
                    <span className="flex size-8 shrink-0 items-center justify-center rounded-full bg-sidebar-primary text-xs font-semibold text-sidebar-primary-foreground">
                        {initials}
                    </span>
                    <span className="grid min-w-0 flex-1 leading-tight">
                        <span className="truncate text-sm font-medium">{name}</span>
                        <span className="truncate text-xs text-muted-foreground">{session.farm.name}</span>
                    </span>
                    <ChevronRight className="ml-auto text-muted-foreground" />
                </SidebarMenuButton>
            </SidebarMenuItem>
        </SidebarMenu>
    );
}

function AppSidebar() {
    const base = useAppBase();

    return (
        <Sidebar variant="inset" collapsible="icon">
            <SidebarHeader>
                {/* A plain link rather than a menu button, so it has no hover/pressed/active highlight */}
                <Link
                    // The demo's logo leads back to the landing page
                    href={base ? "/" : "/dashboard"}
                    className="flex h-12 items-center gap-2.5 rounded-md px-1 outline-none focus-visible:ring-2 focus-visible:ring-sidebar-ring group-data-[collapsible=icon]:px-0"
                >
                    {/*<span className="flex size-8 shrink-0 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">*/}
                    {/*    <Sprout className="size-4.5" />*/}
                    {/*</span>*/}
                    <span className="grid leading-tight group-data-[collapsible=icon]:hidden w-full">
                        <span className="truncate font-logo text-2xl font-semibold text-center">Agropal</span>
                        {/*<span className="truncate text-xs text-muted-foreground font-medium">Farm workspace</span>*/}
                    </span>
                </Link>
            </SidebarHeader>

            <SidebarSeparator className="mx-0" />

            <SidebarContent>
                {sections.map((section) => (
                    <SidebarGroup key={section.label}>
                        <SidebarGroupLabel className="uppercase font-semibold tracking-wider text-[0.7rem]">
                            {section.label}
                        </SidebarGroupLabel>
                        <SidebarGroupContent>
                            <NavMenu items={section.items} />
                        </SidebarGroupContent>
                    </SidebarGroup>
                ))}
            </SidebarContent>

            <SidebarFooter>
                <NavMenu items={secondaryNav} />
                <SidebarMenu>
                    <SidebarMenuItem>
                        <SidebarMenuButton tooltip="Help & support (coming soon)" aria-disabled>
                            <LifeBuoy />
                            <span className="truncate">Help & support</span>
                        </SidebarMenuButton>
                        <SidebarMenuBadge className="text-[0.65rem] text-muted-foreground">Soon</SidebarMenuBadge>
                    </SidebarMenuItem>
                </SidebarMenu>
                <SidebarSeparator className="mx-0" />
                <UserCard />
            </SidebarFooter>

            <SidebarRail />
        </Sidebar>
    );
}

export default AppSidebar;
