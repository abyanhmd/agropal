import React from 'react';
import Link from "next/link";
import {Eye, MessageSquare, MessagesSquare, ThumbsUp} from "lucide-react";
import {PostAuthor, PostSchema} from "@/app/(app)/discussions/_types/post.schema";
import {numberFormat} from "@/lib/format";
import {useAppHref} from "@/lib/demo/base";

export interface PostListProps {
    posts: PostSchema[];
    authors: Map<string, PostAuthor>;
    emptyMessage?: string;
}

// UTC to match how the mock dates are created, so server and browser render the same text
const dateFormat = new Intl.DateTimeFormat("en-US", {month: "short", day: "numeric", year: "numeric", timeZone: "UTC"});
export const formatPostDate = (date: Date | string) => dateFormat.format(new Date(date));

// "Grace Liu" → "GL"
function initials(name: string): string {
    return name.split(/\s+/).filter(Boolean).slice(0, 2).map((part) => part[0]).join("").toUpperCase();
}

export function AuthorAvatar({author, size = "md"}: {author?: PostAuthor, size?: "sm" | "md"}) {
    return (
        <span
            aria-hidden
            className={`flex shrink-0 items-center justify-center rounded-full bg-accent-muted font-sans font-medium text-accent-strong ${size === "sm" ? "size-7 text-xs" : "size-9 text-sm"}`}
        >
            {author ? initials(author.Name) : "?"}
        </span>
    );
}

export function EngagementStat({icon: Icon, value, label}: {icon: typeof Eye, value: number, label: string}) {
    return (
        <span className="inline-flex items-center gap-1 text-sm text-muted-foreground font-sans tabular-nums" title={label}>
            <Icon className="size-4" />
            {numberFormat.format(value)}
            <span className="sr-only">{label}</span>
        </span>
    );
}

export function PostListSkeleton() {
    return (
        <div className="flex flex-col gap-3 animate-pulse">
            {Array.from({length: 6}, (_, i) => (
                <div key={i} className="flex flex-col gap-3 rounded-xl border border-border bg-background p-4">
                    <div className="flex items-center gap-2">
                        <div className="size-9 rounded-full bg-accent-muted" />
                        <div className="h-3 w-40 rounded bg-muted" />
                    </div>
                    <div className="h-5 w-2/3 rounded bg-muted" />
                    <div className="h-3 w-full rounded bg-muted" />
                    <div className="h-3 w-32 rounded bg-muted" />
                </div>
            ))}
        </div>
    );
}

export function PostEmptyState({message}: {message?: string}) {
    return (
        <div className="w-full flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-12 text-muted-foreground">
            <MessagesSquare className="w-8 h-8" />
            <p className="font-sans text-sm">{message ?? "No discussions yet"}</p>
        </div>
    );
}

export default function PostList({posts, authors, emptyMessage}: PostListProps) {
    const appHref = useAppHref();
    if (posts.length === 0) {
        return <PostEmptyState message={emptyMessage} />;
    }

    return (
        <ul className="flex flex-col gap-3">
            {posts.map((post) => {
                const author = authors.get(post.UserID);

                return (
                    // Same pattern as the product cards: the title link stretches over the whole card
                    <li key={post.ID} className="group relative flex flex-col gap-2 rounded-xl border border-border bg-background p-4 transition-colors hover:bg-muted/60 focus-within:bg-muted/60">
                        <div className="flex items-center gap-2 min-w-0">
                            <AuthorAvatar author={author} />
                            <p className="truncate text-sm font-sans">
                                <span className="font-medium">{author?.Name ?? "Unknown author"}</span>
                                <span className="text-muted-foreground"> · {formatPostDate(post.CreatedAt)}</span>
                            </p>
                        </div>

                        <h2 className="font-serif text-lg font-medium">
                            <Link
                                href={appHref(`/discussions/${post.Slug}`)}
                                className="outline-none after:absolute after:inset-0 after:content-[''] group-hover:underline group-focus-within:underline underline-offset-4"
                            >
                                {post.Title}
                            </Link>
                        </h2>

                        {/* First paragraph only; the post page shows the rest */}
                        <p className="line-clamp-2 text-sm text-muted-foreground font-sans">{post.Content.split("\n\n")[0]}</p>

                        <div className="flex flex-wrap items-center gap-4 pt-1">
                            <EngagementStat icon={ThumbsUp} value={post.LikeCount} label="likes" />
                            <EngagementStat icon={MessageSquare} value={post.CommentCount} label="comments" />
                            <EngagementStat icon={Eye} value={post.ViewCount} label="views" />
                        </div>
                    </li>
                );
            })}
        </ul>
    );
}
