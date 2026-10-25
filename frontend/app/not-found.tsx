import type {Metadata} from "next";
import NotFoundContent from "@/app/_components/not-found-content";
import {NOT_FOUND_TITLE} from "@/app/_lib/titles";

// Server wrapper so the 404 page can set its tab title; the page itself is interactive
export const metadata: Metadata = {
    title: NOT_FOUND_TITLE,
};

export default function NotFound() {
    return <NotFoundContent />;
}
