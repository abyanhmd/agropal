import type {Metadata} from "next";
import PostLoader from "@/app/(app)/discussions/_components/loader.discussion";

export const metadata: Metadata = {title: "Discussion"};

export default async function DemoPostPage({params}: {params: Promise<{slug: string}>}) {
    const {slug} = await params;
    return <PostLoader slug={slug} />;
}
