import type {Metadata} from "next";
import DiscussionPage from "@/app/(app)/discussions/page";

export const metadata: Metadata = {title: "Discussions"};

export default function DemoDiscussionsPage() {
    return <DiscussionPage />;
}
