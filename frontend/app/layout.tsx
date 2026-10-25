import type { Metadata } from "next";
import {Fraunces, Geist, Geist_Mono, Plus_Jakarta_Sans, Unbounded} from "next/font/google";
import "./globals.css";
import {TooltipProvider} from "@/components/ui/tooltip";
import {ThemeProvider} from "next-themes";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

const serif  = Fraunces({
  variable: "--font-serif",
  subsets: ["latin"],
})

const sans = Plus_Jakarta_Sans({
  variable: "--font-sans",
  subsets: ["latin"],
});

// Only used for the "Agropal" wordmark, so just the one weight
const logo = Unbounded({
  variable: "--font-logo",
  weight: "600",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  // Pages set just their own name, e.g. "Products" → "Products · Agropal"
  title: {
    template: "%s · Agropal",
    default: "Agropal",
  },
  description: "Track your farm's products, orders, discussions, and weather in one place.",
};

export default function RootLayout({ children }: LayoutProps<"/">) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} ${serif.variable} ${sans.variable} ${logo.variable} h-full antialiased`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col">
        <ThemeProvider attribute="class" defaultTheme="system" enableSystem disableTransitionOnChange>
          <TooltipProvider>
            {children}
          </TooltipProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
