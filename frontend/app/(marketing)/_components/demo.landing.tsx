import React from 'react';
import Image from "next/image";
import Link from "next/link";
import {ArrowRight, CalendarDays, CloudSun, LayoutDashboard, MessagesSquare, Package, ShoppingCart} from "lucide-react";

// Doors into the public demo: the real app pages, filled with a season's worth of sample data

interface DemoPage {
    title: string;
    description: string;
    href: string;
    icon: React.ComponentType<{className?: string}>;
    photo: string;
}

const PAGES: DemoPage[] = [
    {title: "Dashboard", description: "Stock in and out, month by month, and what's selling.", href: "/demo/dashboard", icon: LayoutDashboard, photo: "/demo/tomatoes.jpg"},
    {title: "Products", description: "Ten crops with photos, prices and stock history.", href: "/demo/products", icon: Package, photo: "/demo/eggs.jpg"},
    {title: "Orders", description: "A year of sales and expenses, tied to each product.", href: "/demo/orders", icon: ShoppingCart, photo: "/demo/carrots.jpg"},
    {title: "Discussions", description: "Growers swapping advice on pests, prices and planting.", href: "/demo/discussions", icon: MessagesSquare, photo: "/demo/lettuce.jpg"},
    {title: "Crop calendar", description: "What to sow and harvest this month, for your location.", href: "/demo/calendar", icon: CalendarDays, photo: "/demo/peppers.jpg"},
    {title: "Weather", description: "A forecast for the field, with soil temperature.", href: "/demo/weather", icon: CloudSun, photo: "/demo/wheat.jpg"},
];

export default function DemoSection() {
    return (
        <section id="demo" className="scroll-mt-16 mx-auto flex w-full max-w-6xl flex-col gap-10 px-4 py-20">
            <div className="flex max-w-2xl flex-col gap-3">
                <h2 className="font-serif text-3xl font-medium sm:text-4xl">See it after a season of use</h2>
                <p className="font-sans text-muted-foreground">
                    Look around Maple Hollow Farm, a small mixed vegetable farm, with its harvests, sales and conversations
                    already in. No account needed, and nothing you click is saved.
                </p>
            </div>

            <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                {PAGES.map((page) => (
                    <Link
                        key={page.href}
                        href={page.href}
                        className="group flex flex-col overflow-hidden rounded-2xl border border-border bg-background transition-shadow hover:shadow-lg hover:shadow-black/5"
                    >
                        <div className="relative aspect-[16/10] overflow-hidden bg-muted">
                            <Image
                                src={page.photo}
                                alt=""
                                fill
                                sizes="(min-width: 1024px) 384px, (min-width: 640px) 50vw, 100vw"
                                className="object-cover transition-transform duration-300 group-hover:scale-105"
                            />
                        </div>
                        <div className="flex flex-1 flex-col gap-1.5 p-5">
                            <span className="flex items-center gap-2 font-serif text-xl">
                                <page.icon className="size-4 text-muted-foreground" />
                                {page.title}
                            </span>
                            <p className="font-sans text-sm text-muted-foreground">{page.description}</p>
                            <span className="mt-auto flex items-center gap-1 pt-3 font-sans text-sm text-accent-strong">
                                Open <ArrowRight className="size-4 transition-transform group-hover:translate-x-0.5" />
                            </span>
                        </div>
                    </Link>
                ))}
            </div>

            <p className="font-sans text-xs text-muted-foreground">
                Sample photos from Wikimedia Commons, used under their open licences.{" "}
                <a href="/demo/CREDITS.md" className="underline underline-offset-4 hover:text-foreground">Photo credits</a>
            </p>
        </section>
    );
}
