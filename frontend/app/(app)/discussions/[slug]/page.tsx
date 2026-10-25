import React from 'react';
import type {Metadata} from "next";
import PostLoader from "@/app/(app)/discussions/_components/loader.discussion";

// The post's own title replaces this once it renders (see PostDetail). Posts load in the browser,
// since they need the signed-in user's token.
export const metadata: Metadata = {
    title: "Discussion",
};

export default async function PostDetailPage({params}: PageProps<"/discussions/[slug]">) {
    const {slug} = await params;
    return <PostLoader slug={slug} />;
}
