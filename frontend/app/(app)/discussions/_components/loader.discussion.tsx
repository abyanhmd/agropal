'use client';

import React from 'react';
import {usePost} from "@/lib/api/hooks";
import QueryError from "@/components/app/query-error";
import NotFoundState from "@/components/app/not-found-state";
import PostDetail, {PostDetailSkeleton} from "@/app/(app)/discussions/_components/detail.discussion";
import {useAppHref} from "@/lib/demo/base";

// Loads a post by its slug
export default function PostLoader({slug}: {slug: string}) {
    const appHref = useAppHref();
    const {data: detail, error, reload} = usePost(slug);
    if (error && detail === undefined) return <div className="page-container"><QueryError error={error} onRetry={reload} /></div>;
    if (detail === undefined) return <PostDetailSkeleton />;
    if (detail === null) return <NotFoundState what="discussion" backHref={appHref("/discussions")} />;

    return <PostDetail detail={detail} />;
}
