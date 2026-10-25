'use client';

import React, {useEffect, useMemo, useState} from 'react';
import Link from "next/link";
import {ArrowLeft, Eye, MessageSquare, Reply} from "lucide-react";
import {PostAuthor, PostCommentSchema, PostDetailSchema} from "@/app/(app)/discussions/_types/post.schema";
import {AuthorAvatar, EngagementStat, formatPostDate} from "@/app/(app)/discussions/_components/list.discussion";
import ReactionButtons, {Reaction} from "@/app/(app)/discussions/_components/reactions.discussion";
import CommentForm from "@/app/(app)/discussions/_components/comment-form.discussion";
import {useSetBreadcrumb} from "@/hooks/breadcrumb-state";
import {useAuth} from "@/lib/auth/auth-provider";
import {addPostComment, reactTo, recordView} from "@/lib/api/hooks";
import {authorsFrom} from "@/app/(app)/discussions/_lib/authors";
import {useAppHref} from "@/lib/demo/base";

export interface PostDetailProps {
    detail: PostDetailSchema;
}

interface ThreadActions {
    // Counts and the viewer's own reaction, with a reaction just clicked already applied
    reactionOf: (id: string, item: Reacted) => {likes: number, dislikes: number, value: Reaction};
    react: (target: "post" | "comment", assetID: string, value: Reaction) => void;
    replyingTo: string | null;
    setReplyingTo: (commentID: string | null) => void;
    addReply: (content: string, parentID: string) => Promise<void>;
}

interface Reacted {
    LikeCount: number;
    DislikeCount: number;
    ViewerReaction?: "LIKE" | "DISLIKE";
}

interface CommentNode {
    comment: PostCommentSchema;
    replies: CommentNode[];
}

// Flat list → tree via ParentID; a reply whose parent is missing is shown at the top level
function buildThread(comments: PostCommentSchema[]): CommentNode[] {
    const nodes = new Map(comments.map((comment) => [comment.ID, {comment, replies: []} as CommentNode]));
    const roots: CommentNode[] = [];

    for (const comment of comments) {
        const node = nodes.get(comment.ID)!;
        const parent = comment.ParentID ? nodes.get(comment.ParentID) : undefined;
        (parent ? parent.replies : roots).push(node);
    }

    return roots;
}

// Deeper replies stop indenting so long threads don't squeeze off narrow screens:
// past this level, all further replies are listed flat under the last indented comment
const MAX_INDENT = 3;

interface VisibleReply {
    node: CommentNode;
    replyTo?: string; // shown when the reply answers someone other than the comment it's listed under
}

// Every reply below a node, oldest first, each tagged with whose comment it answers
function flattenReplies(node: CommentNode, authors: Map<string, PostAuthor>): VisibleReply[] {
    const result: VisibleReply[] = [];
    const walk = (parent: CommentNode) => parent.replies.forEach((reply) => {
        result.push({
            node: reply,
            replyTo: parent === node ? undefined : authors.get(parent.comment.UserID)?.Name ?? "Unknown author",
        });
        walk(reply);
    });
    walk(node);

    return result.sort((a, b) => new Date(a.node.comment.CreatedAt).getTime() - new Date(b.node.comment.CreatedAt).getTime());
}

// Thread lines, YouTube style: a line runs down from the parent's avatar and curves into each
// reply's avatar. The geometry below assumes the size="sm" avatar (28px), the 12px gaps (gap-3)
// between a comment and its replies, and the 24px reply indent (pl-6).
// Every segment sits at the same whole-pixel x (13px, under the avatar's center), so the joins
// line up exactly; a centered 1px line would land on 13.5px and blur or step at each join.
const THREAD_LINE = "border-border";

function CommentItem({node, authors, depth, actions, isLast = true, flat = false, replyTo}: {
    node: CommentNode,
    authors: Map<string, PostAuthor>,
    depth: number,
    actions: ThreadActions,
    isLast?: boolean,  // last sibling: the line ends at this reply's curve instead of running past it
    flat?: boolean,    // replies are listed by the parent instead of nested here
    replyTo?: string,
}) {
    const {comment} = node;
    const author = authors.get(comment.UserID);
    const isReplying = actions.replyingTo === comment.ID;

    const children: VisibleReply[] = flat
        ? []
        : depth + 1 >= MAX_INDENT
            ? flattenReplies(node, authors)
            : node.replies.map((reply) => ({node: reply}));

    return (
        <li className={`relative flex flex-col gap-3 ${depth > 0 ? "pl-6" : ""}`}>
            {depth > 0 &&
                <>
                    {/* Curve from the parent's line into this avatar's center; starts in the gap above */}
                    <span aria-hidden className={`absolute -top-3 left-0 h-[27px] w-[22px] rounded-bl-2xl border-b border-l ${THREAD_LINE}`} />
                    {/* Carries the parent's line straight past this reply to the next sibling.
                        It starts at the top, not at the curve's end, since the curve bends away well above that */}
                    {!isLast && <span aria-hidden className={`absolute bottom-0 left-0 top-0 border-l ${THREAD_LINE}`} />}
                </>
            }

            <div className="flex gap-3">
                <div className="flex shrink-0 flex-col">
                    <AuthorAvatar author={author} size="sm" />
                    {/* Starts this comment's own line, down to its first reply's curve */}
                    {children.length > 0 && <span aria-hidden className={`-mb-3 ml-[13px] mt-1 flex-1 border-l ${THREAD_LINE}`} />}
                </div>
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <p className="text-sm font-sans">
                        <span className="font-medium">{author?.Name ?? "Unknown author"}</span>
                        <span className="text-muted-foreground"> · {formatPostDate(comment.CreatedAt)}</span>
                    </p>
                    {replyTo && <p className="text-xs text-muted-foreground font-sans">Replying to {replyTo}</p>}
                    <p className="whitespace-pre-line break-words text-sm font-sans">{comment.Content}</p>
                    <div className="flex flex-wrap items-center gap-2 pt-0.5">
                        <ReactionButtons
                            size="sm"
                            {...actions.reactionOf(comment.ID, comment)}
                            onChange={(value) => actions.react("comment", comment.ID, value)}
                        />
                        <button
                            type="button"
                            aria-expanded={isReplying}
                            onClick={() => actions.setReplyingTo(isReplying ? null : comment.ID)}
                            className="inline-flex h-7 cursor-pointer items-center gap-1 rounded-full px-2 text-xs font-sans text-muted-foreground outline-none transition-colors hover:bg-muted hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50"
                        >
                            <Reply className="size-3.5" />
                            Reply
                        </button>
                    </div>

                    {isReplying &&
                        <div className="pt-1">
                            <CommentForm
                                autoFocus
                                placeholder={`Reply to ${author?.Name ?? "this comment"}…`}
                                submitLabel="Reply"
                                onSubmit={(content) => actions.addReply(content, comment.ID)}
                                onCancel={() => actions.setReplyingTo(null)}
                            />
                        </div>
                    }
                </div>
            </div>

            {children.length > 0 &&
                // Left edge sits on the avatar's line (13px), so each reply's curve continues it
                <ul className="ml-[13px] flex flex-col gap-3">
                    {children.map((child, i) => (
                        <CommentItem
                            key={child.node.comment.ID}
                            node={child.node}
                            authors={authors}
                            depth={depth + 1}
                            actions={actions}
                            isLast={i === children.length - 1}
                            flat={depth + 1 >= MAX_INDENT}
                            replyTo={child.replyTo}
                        />
                    ))}
                </ul>
            }
        </li>
    );
}

export function PostDetailSkeleton() {
    return (
        <div className="page-container flex animate-pulse flex-col gap-6">
            <div className="h-4 w-32 rounded bg-muted" />
            <div className="flex flex-col gap-3">
                <div className="h-8 w-3/4 rounded bg-muted" />
                <div className="flex items-center gap-2">
                    <div className="size-9 rounded-full bg-accent-muted" />
                    <div className="h-3 w-40 rounded bg-muted" />
                </div>
            </div>
            <div className="flex flex-col gap-2">
                {Array.from({length: 4}, (_, i) => <div key={i} className="h-3 w-full rounded bg-muted" />)}
            </div>
            <div className="h-48 rounded-xl bg-muted" />
        </div>
    );
}

export default function PostDetail({detail}: PostDetailProps) {
    const appHref = useAppHref();
    const {Post: post} = detail;
    useSetBreadcrumb([{label: "Discussion", href: "/discussions"}, {label: post.Title}]);

    // Tab title: the post's title, set here because the route's metadata can't depend on the slug
    useEffect(() => {
        document.title = `${post.Title} · Agropal`;
    }, [post.Title]);

    const {session} = useAuth();
    const viewerID = session.status === "ready" ? session.user.user_id : "";

    // A clicked reaction shows at once, before the server has answered. Once the page reloads
    // (a new detail), the server's answer replaces it.
    const [clicked, setClicked] = useState<{source: PostDetailSchema, values: Record<string, Reaction>}>({source: detail, values: {}});
    const clickedValues = clicked.source === detail ? clicked.values : {};
    const [busy, setBusy] = useState<string[]>([]);

    useEffect(() => {
        recordView(post.ID);
    }, [post.ID]);

    const [replyingTo, setReplyingTo] = useState<string | null>(null);

    const comments = detail.Comments;
    const viewerName = session.status === "ready" ? session.user.name.trim() || session.user.username : "";
    const authors = useMemo(() => {
        const known = authorsFrom([post, ...comments]);
        // Someone who hasn't posted yet still needs a name on their comment box
        if (viewerID && !known.some((a) => a.ID === viewerID)) known.push({ID: viewerID, Name: viewerName});
        return new Map(known.map((a) => [a.ID, a]));
    }, [post, comments, viewerID, viewerName]);
    const thread = useMemo(() => buildThread(comments), [comments]);
    const author = authors.get(post.UserID);

    const actions: ThreadActions = {
        reactionOf: (id, item) => {
            const value = id in clickedValues ? clickedValues[id] : item.ViewerReaction ?? null;
            const was = item.ViewerReaction ?? null;
            return {
                value,
                likes: item.LikeCount - (was === "LIKE" ? 1 : 0) + (value === "LIKE" ? 1 : 0),
                dislikes: item.DislikeCount - (was === "DISLIKE" ? 1 : 0) + (value === "DISLIKE" ? 1 : 0),
            };
        },
        react: (target, assetID, value) => {
            // One request per item at a time, so a quick double click can't land out of order
            if (busy.includes(assetID)) return;
            setBusy((ids) => [...ids, assetID]);
            setClicked({source: detail, values: {...clickedValues, [assetID]: value}});
            reactTo(target, assetID, value)
                .catch(() => {
                    // Not saved: go back to what the server has
                    const rest = {...clickedValues};
                    delete rest[assetID];
                    setClicked({source: detail, values: rest});
                })
                .finally(() => setBusy((ids) => ids.filter((id) => id !== assetID)));
        },
        replyingTo,
        setReplyingTo,
        addReply: async (content, parentID) => {
            await addPostComment(post.ID, content, parentID);
            setReplyingTo(null);
        },
    };

    return (
        <div className="page-container flex flex-col gap-3">
            <Link
                href={appHref("/discussions")}
                className="inline-flex w-fit items-center gap-1.5 text-sm text-muted-foreground font-sans transition-colors hover:text-foreground"
            >
                <ArrowLeft className="size-4" />
                Back to discussion
            </Link>

            <article className="flex flex-col gap-4">
                <h1 className="font-serif text-3xl font-medium">{post.Title}</h1>

                <div className="flex items-center gap-2">
                    <AuthorAvatar author={author} />
                    <p className="text-sm font-sans">
                        <span className="font-medium">{author?.Name ?? "Unknown author"}</span>
                        <span className="text-muted-foreground">
                            {" · "}{formatPostDate(post.CreatedAt)}
                            {post.UpdatedAt && ` · edited ${formatPostDate(post.UpdatedAt)}`}
                        </span>
                    </p>
                </div>

                <div className="flex flex-col gap-3 font-sans leading-relaxed">
                    {post.Content.split("\n\n").map((paragraph, i) => <p key={i}>{paragraph}</p>)}
                </div>

                <div className="flex flex-wrap items-center gap-4 border-b border-border pb-3">
                    <ReactionButtons
                        {...actions.reactionOf(post.ID, post)}
                        onChange={(value) => actions.react("post", post.ID, value)}
                    />
                    <EngagementStat icon={MessageSquare} value={post.CommentCount} label="comments" />
                    <EngagementStat icon={Eye} value={post.ViewCount} label="views" />
                </div>
            </article>

            <section className="flex flex-col gap-4">
                <h2 className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
                    {comments.length === 1 ? "1 comment" : `${comments.length} comments`}
                </h2>

                <div className="flex gap-3">
                    <AuthorAvatar author={authors.get(viewerID)} size="sm" />
                    <div className="min-w-0 flex-1">
                        <CommentForm onSubmit={(content) => addPostComment(post.ID, content)} />
                    </div>
                </div>

                {thread.length > 0
                    ? <ul className="flex flex-col gap-5">
                        {thread.map((node) => (
                            <CommentItem key={node.comment.ID} node={node} authors={authors} depth={0} actions={actions} />
                        ))}
                    </ul>
                    : <div className="flex flex-col items-center gap-2 rounded-xl border border-dashed border-border py-10 text-muted-foreground">
                        <MessageSquare className="size-7" />
                        <p className="font-sans text-sm">No comments yet. Start the conversation below.</p>
                    </div>
                }
            </section>
        </div>
    );
}
