import React from 'react';
import Link from "next/link";
import Logo from "@/app/(marketing)/_components/logo.landing";
import ThemeToggle from "@/app/(marketing)/_components/theme-toggle.landing";
import GetStartedButton from "@/app/(marketing)/_components/get-started-button.landing";
import GithubIcon, {GITHUB_URL} from "@/app/(marketing)/_components/github-icon.landing";
import {buttonVariants} from "@/components/ui/button";

export default function Header() {
    return (
        <header className="sticky top-0 z-20 border-b border-border/60 bg-background/80 backdrop-blur">
            <div className="mx-auto flex h-16 w-full max-w-6xl items-center justify-between gap-4 px-4">
                <Logo />
                {/* "/#…" rather than "#…", so the links also work from other marketing pages */}
                <nav className="hidden items-center gap-6 font-sans text-sm text-muted-foreground sm:flex">
                    <Link href="/#features" className="transition-colors hover:text-foreground">Features</Link>
                    <Link href="/#demo" className="transition-colors hover:text-foreground">Demo</Link>
                    <Link href="/#how-it-works" className="transition-colors hover:text-foreground">How it works</Link>
                </nav>
                <div className="flex items-center gap-2">
                    <a
                        href={GITHUB_URL}
                        target="_blank"
                        rel="noopener noreferrer"
                        aria-label="AgroPal on GitHub"
                        className={`${buttonVariants({variant: "ghost", size: "icon"})} rounded-xl`}
                    >
                        <GithubIcon className="size-5" />
                    </a>
                    <ThemeToggle />
                    <Link href="/login" className="hidden px-2 font-sans text-sm text-muted-foreground transition-colors hover:text-foreground sm:inline">Sign in</Link>
                    <GetStartedButton size="default">Get started</GetStartedButton>
                </div>
            </div>
        </header>
    );
}
