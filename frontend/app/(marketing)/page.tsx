import React from 'react';
import Link from "next/link";
import {ArrowRight, CloudSun, MapPin, MessagesSquare, Package, Receipt, ShoppingCart, Sprout} from "lucide-react";
import {Button} from "@/components/ui/button";
import LandingPreview from "@/app/(marketing)/_components/preview.landing";
import GetStartedButton from "@/app/(marketing)/_components/get-started-button.landing";
import DemoSection from "@/app/(marketing)/_components/demo.landing";

// The public landing page (header and footer come from the (marketing) layout); the app lives under /dashboard

interface Feature {
    title: string;
    description: string;
    icon: React.ComponentType<{className?: string}>;
    // Each feature card picks up one of the palette's soft tones
    tone: string;
}

const FEATURES: Feature[] = [
    {
        title: "Products & stock",
        description: "Log every harvest coming in and every sale going out. See what's piling up and what's running low, per unit.",
        icon: Package,
        tone: "bg-accent-muted text-accent-strong",
    },
    {
        title: "Orders",
        description: "Keep income and expenses next to the products they belong to, so you know which crops actually pay.",
        icon: ShoppingCart,
        tone: "bg-warning-muted text-warning-strong",
    },
    {
        title: "Discussions",
        description: "Ask other growers about pests, prices, or planting dates, and keep the answers in one searchable place.",
        icon: MessagesSquare,
        tone: "bg-info-muted text-info-strong",
    },
    {
        title: "Weather",
        description: "A forecast for your own field, not the nearest city, including soil temperature and moisture.",
        icon: CloudSun,
        tone: "bg-danger-muted text-danger-strong",
    },
];

const STEPS = [
    {title: "Pin your land", description: "Drop a pin on the map so the forecast is for your field.", icon: MapPin},
    {title: "Add your products", description: "List what you grow and record stock as it moves.", icon: Sprout},
    {title: "Track the money", description: "Attach income and expenses to each product as orders.", icon: Receipt},
];

function Hero() {
    return (
        <section className="mx-auto grid w-full max-w-6xl items-center gap-12 px-4 py-16 sm:py-24 lg:grid-cols-[1.1fr_1fr]">
            <div className="flex flex-col items-start gap-6">
                <span className="flex items-center gap-1.5 rounded-full bg-accent-muted px-3 py-1 font-sans text-xs font-medium text-accent-strong">
                    <Sprout className="size-3.5" />
                    For small farms and growers
                </span>
                <h1 className="font-serif text-4xl font-medium leading-tight tracking-tight sm:text-5xl lg:text-6xl">
                    Your whole farm,<br />in one tidy ledger.
                </h1>
                <p className="max-w-lg font-sans text-lg text-muted-foreground">
                    Agropal keeps your stock, orders, weather, and conversations together, so you spend less time
                    on notebooks and spreadsheets and more time in the field.
                </p>
                <div className="flex flex-wrap items-center gap-3">
                    <GetStartedButton>
                        Get started
                        <ArrowRight />
                    </GetStartedButton>
                    <Button size="lg" variant="outline" nativeButton={false} render={<Link href="/demo/dashboard" />} className="rounded-xl px-4">
                        Open dashboard
                    </Button>
                </div>
            </div>

            <LandingPreview />
        </section>
    );
}

function Features() {
    return (
        <section id="features" className="scroll-mt-16 border-y border-border bg-muted/60">
            <div className="mx-auto flex w-full max-w-6xl flex-col gap-10 px-4 py-20">
                <div className="flex max-w-2xl flex-col gap-3">
                    <h2 className="font-serif text-3xl font-medium sm:text-4xl">Everything the season throws at you</h2>
                    <p className="font-sans text-muted-foreground">
                        Four simple tools that share the same products, so nothing has to be entered twice.
                    </p>
                </div>

                <div className="grid gap-4 sm:grid-cols-2">
                    {FEATURES.map((feature) => (
                        <div key={feature.title} className="flex gap-4 rounded-2xl border border-border bg-background p-6">
                            <span className={`flex size-11 shrink-0 items-center justify-center rounded-xl ${feature.tone}`}>
                                <feature.icon className="size-5" />
                            </span>
                            <div className="flex flex-col gap-1.5">
                                <h3 className="font-serif text-xl">{feature.title}</h3>
                                <p className="font-sans text-sm leading-relaxed text-muted-foreground">{feature.description}</p>
                            </div>
                        </div>
                    ))}
                </div>
            </div>
        </section>
    );
}

function HowItWorks() {
    return (
        <section id="how-it-works" className="scroll-mt-16 mx-auto flex w-full max-w-6xl flex-col gap-10 px-4 py-20">
            <div className="flex max-w-2xl flex-col gap-3">
                <h2 className="font-serif text-3xl font-medium sm:text-4xl">Up and running before lunch</h2>
                <p className="font-sans text-muted-foreground">No setup wizard, no training. Three steps and you&apos;re tracking.</p>
            </div>

            {/* A vertical timeline on phones (number beside the text), three columns from sm up */}
            <ol className="grid sm:grid-cols-3 sm:gap-6">
                {STEPS.map((step, i) => {
                    const isLast = i === STEPS.length - 1;
                    return (
                        <li key={step.title} className="flex gap-4 sm:flex-col sm:gap-3">
                            <div className="flex flex-col items-center sm:flex-row sm:gap-3">
                                <span className="flex size-10 shrink-0 items-center justify-center rounded-full border border-border bg-background font-serif text-lg text-accent-strong">
                                    {i + 1}
                                </span>
                                {/* Dashed "furrow" to the next step: down on phones, across from sm up */}
                                {!isLast &&
                                    <span className="my-2 w-px flex-1 border-l border-dashed border-border sm:my-0 sm:h-px sm:w-auto sm:border-l-0 sm:border-t" />
                                }
                            </div>
                            {/* pt centers the title on the circle when they sit side by side */}
                            <div className={`flex flex-col gap-1.5 pt-1.5 sm:gap-3 sm:pt-0 ${isLast ? "" : "pb-8 sm:pb-0"}`}>
                                <div className="flex items-center gap-2">
                                    <step.icon className="size-4 text-muted-foreground" />
                                    <h3 className="font-serif text-xl">{step.title}</h3>
                                </div>
                                <p className="font-sans text-sm text-muted-foreground">{step.description}</p>
                            </div>
                        </li>
                    );
                })}
            </ol>
        </section>
    );
}

function CallToAction() {
    return (
        <section className="mx-auto w-full max-w-6xl px-4 pb-20">
            <div className="flex flex-col items-start gap-6 rounded-3xl bg-accent-muted px-8 py-12 sm:flex-row sm:items-center sm:justify-between sm:px-12">
                <div className="flex flex-col gap-2">
                    <h2 className="font-serif text-3xl font-medium">Ready to clear off the kitchen table?</h2>
                    <p className="font-sans text-muted-foreground">Create a free account and start with your first product.</p>
                </div>
                <GetStartedButton>
                    Get started
                    <ArrowRight />
                </GetStartedButton>
            </div>
        </section>
    );
}

export default function LandingPage() {
    return (
        <>
            <Hero />
            <Features />
            <DemoSection />
            <HowItWorks />
            <CallToAction />
        </>
    );
}
