'use client';

import React from 'react';
import {useTheme} from "next-themes";
import {Moon, Sun} from "lucide-react";
import {Button} from "@/components/ui/button";

// Quick light/dark switch for the landing header; the app picks themes in Settings
export default function ThemeToggle() {
    const {resolvedTheme, setTheme} = useTheme();

    return (
        <Button
            variant="ghost"
            size="icon"
            aria-label="Toggle theme"
            className="relative cursor-pointer rounded-xl"
            onClick={() => setTheme(resolvedTheme === "dark" ? "light" : "dark")}
        >
            {/* Both icons render and CSS picks one, so server and client markup match */}
            <Sun className="size-5 scale-100 rotate-0 transition-transform dark:scale-0 dark:-rotate-90" />
            <Moon className="absolute size-5 scale-0 rotate-90 transition-transform dark:scale-100 dark:rotate-0" />
        </Button>
    );
}
