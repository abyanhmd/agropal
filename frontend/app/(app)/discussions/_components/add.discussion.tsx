'use client';

import React, {useState} from 'react';
import {useRouter} from "next/navigation";
import {Plus} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger} from "@/components/ui/dialog";
import {Input} from "@/components/ui/input";
import {Label} from "@/components/ui/label";
import {addPost} from "@/lib/api/hooks";
import {errorMessage} from "@/app/(auth)/_components/auth-form";
import {useAppHref} from "@/lib/demo/base";

// Form state lives in its own component, so it starts fresh each time the dialog opens
function AddPostForm({onDone}: {onDone: (slug: string) => void}) {
    const [title, setTitle] = useState("");
    const [content, setContent] = useState("");
    const [pending, setPending] = useState(false);
    const [error, setError] = useState<string | null>(null);

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setPending(true);
        setError(null);
        try {
            onDone(await addPost(title.trim(), content.trim()));
        } catch (err) {
            setError(errorMessage(err));
            setPending(false);
        }
    };

    return (
        <form onSubmit={handleSubmit} className="grid gap-6">
            <DialogHeader>
                <DialogTitle>New discussion</DialogTitle>
                <DialogDescription>Visible to everyone on Agropal.</DialogDescription>
            </DialogHeader>

            <div className="grid gap-4 font-sans">
                <div className="flex flex-col gap-2">
                    <Label htmlFor="post-title">Title</Label>
                    <Input id="post-title" required autoFocus placeholder="Best cover crops for clay soil?" value={title} onChange={(e) => setTitle(e.target.value)} />
                </div>
                <div className="flex flex-col gap-2">
                    <Label htmlFor="post-content">Details</Label>
                    <textarea
                        id="post-content"
                        required
                        rows={6}
                        value={content}
                        onChange={(e) => setContent(e.target.value)}
                        className="w-full min-w-0 resize-y rounded-xl border border-input bg-transparent px-3 py-2 text-base shadow-xs outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 md:text-sm dark:bg-input/30"
                    />
                </div>
            </div>

            {error && <p role="alert" className="text-sm text-danger-strong font-sans">{error}</p>}

            <DialogFooter showCloseButton>
                <Button type="submit" className="cursor-pointer" disabled={!title.trim() || !content.trim() || pending}>{pending ? "Posting…" : "Post"}</Button>
            </DialogFooter>
        </form>
    );
}

export default function AddPostDialog() {
    const appHref = useAppHref();
    const router = useRouter();
    const [open, setOpen] = useState(false);

    return (
        <Dialog open={open} onOpenChange={setOpen}>
            <DialogTrigger render={<Button className="h-9 w-full cursor-pointer rounded-xl sm:w-auto" />}>
                <Plus />
                New discussion
            </DialogTrigger>
            <DialogContent>
                <AddPostForm
                    onDone={(slug) => {
                        setOpen(false);
                        router.push(appHref(`/discussions/${slug}`));
                    }}
                />
            </DialogContent>
        </Dialog>
    );
}
