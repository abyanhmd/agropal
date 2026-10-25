import React from 'react';
import {ThumbsDown, ThumbsUp} from "lucide-react";
import {numberFormat} from "@/lib/format";

// Mirrors the backend engagement actions; null = no reaction
export type Reaction = "LIKE" | "DISLIKE" | null;

export interface ReactionButtonsProps {
    likes: number;    // counts including the viewer's own reaction
    dislikes: number;
    value: Reaction;
    onChange: (value: Reaction) => void;
    size?: "sm" | "md";
}

const OPTIONS = [
    {reaction: "LIKE", label: "Like", icon: ThumbsUp, active: "border-accent/40 bg-accent-muted text-accent-strong"},
    {reaction: "DISLIKE", label: "Dislike", icon: ThumbsDown, active: "border-danger/40 bg-danger-muted text-danger-strong"},
] as const;

// Like and dislike are exclusive; pressing the active one again takes the reaction back
export default function ReactionButtons({likes, dislikes, value, onChange, size = "md"}: ReactionButtonsProps) {
    return (
        <div role="group" aria-label="Reactions" className="inline-flex items-center gap-1.5">
            {OPTIONS.map(({reaction, label, icon: Icon, active}) => {
                const pressed = value === reaction;
                const count = reaction === "LIKE" ? likes : dislikes;

                return (
                    <button
                        key={reaction}
                        type="button"
                        aria-pressed={pressed}
                        aria-label={`${label} (${count})`}
                        onClick={() => onChange(pressed ? null : reaction)}
                        className={`inline-flex cursor-pointer items-center gap-1 rounded-full border font-sans tabular-nums outline-none transition-colors focus-visible:ring-3 focus-visible:ring-ring/50 ${size === "sm" ? "h-7 px-2 text-xs" : "h-8 px-3 text-sm"} ${pressed ? active : "border-border text-muted-foreground hover:bg-muted hover:text-foreground"}`}
                    >
                        <Icon className={size === "sm" ? "size-3.5" : "size-4"} />
                        {numberFormat.format(count)}
                    </button>
                );
            })}
        </div>
    );
}
