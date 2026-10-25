import React from 'react';
import Link from "next/link";
import {Button} from "@/components/ui/button";

// The landing page's main call to action, used in the header, hero, and closing banner
export default function GetStartedButton({size = "lg", children}: {size?: "default" | "lg", children: React.ReactNode}) {
    return (
        <Button size={size} nativeButton={false} render={<Link href="/register" />} className="rounded-xl px-4">
            {children}
        </Button>
    );
}
