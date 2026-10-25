// Mirrors the backend post entities (PostResponse / PostCommentResponse)

export interface PostSchema {
    ID: string,
    UserID: string,
    Title: string,
    Content: string,
    Slug: string,
    LikeCount: number,
    DislikeCount: number,
    ViewCount: number,
    CommentCount: number,
    AuthorName: string,
    ViewerReaction?: "LIKE" | "DISLIKE", // the signed-in user's own; only set when a post is fetched on its own
    CreatedAt: Date,
    UpdatedAt?: Date,
}

export type PostSchemaType = PostSchema;

export interface PostCommentSchema {
    ID: string,
    PostID: string,
    UserID: string,
    ParentID?: string, // set on replies; points at the comment being answered
    Content: string,
    ReplyCount: number,
    LikeCount: number,
    DislikeCount: number,
    ViewCount: number,
    AuthorName: string,
    ViewerReaction?: "LIKE" | "DISLIKE",
    CreatedAt: Date,
    UpdatedAt?: Date,
}

export type PostCommentSchemaType = PostCommentSchema;

export interface PostDetailSchema {
    Post: PostSchema,
    Comments: PostCommentSchema[],
}

export type PostDetailSchemaType = PostDetailSchema;

// Posts and comments carry their author's id and display name; nothing else about a user is public
export interface PostAuthor {
    ID: string,
    Name: string,
}
