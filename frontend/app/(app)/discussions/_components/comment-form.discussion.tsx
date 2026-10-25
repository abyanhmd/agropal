import React, {useId, useState} from 'react';
import {Button} from "@/components/ui/button";
import {errorMessage} from "@/app/(auth)/_components/auth-form";

export interface CommentFormProps {
    onSubmit: (content: string) => void | Promise<void>;
    onCancel?: () => void; // shown for replies, so the inline form can be closed
    placeholder?: string;
    submitLabel?: string;
    autoFocus?: boolean;
}

const MAX_LENGTH = 2000;

export default function CommentForm({onSubmit, onCancel, placeholder = "Add a comment…", submitLabel = "Comment", autoFocus}: CommentFormProps) {
    const [content, setContent] = useState("");
    const id = useId();
    const trimmed = content.trim();
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const submit = async (e?: React.FormEvent) => {
        e?.preventDefault();
        if (!trimmed || pending) return;
        setPending(true);
        setError(null);
        try {
            await onSubmit(trimmed);
            setContent("");
        } catch (err) {
            // The text stays, so nothing typed is lost
            setError(errorMessage(err));
        } finally {
            setPending(false);
        }
    };

    return (
        <form onSubmit={submit} className="flex flex-col gap-2">
            <label htmlFor={id} className="sr-only">{placeholder}</label>
            <textarea
                id={id}
                value={content}
                onChange={(e) => setContent(e.target.value)}
                onKeyDown={(e) => {
                    // Ctrl/Cmd + Enter posts; plain Enter adds a new line
                    if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) submit();
                    if (e.key === "Escape" && onCancel) onCancel();
                }}
                placeholder={placeholder}
                maxLength={MAX_LENGTH}
                rows={3}
                autoFocus={autoFocus}
                className="w-full min-w-0 resize-y rounded-xl border border-input bg-transparent px-3 py-2 text-base font-sans shadow-xs transition-[color,box-shadow] outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 md:text-sm dark:bg-input/30"
            />
            <div className="flex items-center justify-between gap-2">
                <span className="text-xs text-muted-foreground font-sans">Ctrl + Enter to post</span>
                <div className="flex gap-2">
                    {onCancel && <Button type="button" variant="ghost" size="sm" onClick={onCancel}>Cancel</Button>}
                    <Button type="submit" size="sm" disabled={!trimmed || pending}>{pending ? "Posting…" : submitLabel}</Button>
                </div>
            </div>
            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}
        </form>
    );
}
