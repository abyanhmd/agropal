import {PostAuthor} from "@/app/(app)/discussions/_types/post.schema";

// Posts and comments carry their author's id and display name, which is all anyone can see of
// another user. Authors with no name on record are left out and shown as unknown.
export function authorsFrom(items: {UserID: string, AuthorName: string}[]): PostAuthor[] {
    const authors = new Map<string, PostAuthor>();
    for (const {UserID, AuthorName} of items) {
        if (AuthorName && !authors.has(UserID)) authors.set(UserID, {ID: UserID, Name: AuthorName});
    }
    return [...authors.values()];
}
