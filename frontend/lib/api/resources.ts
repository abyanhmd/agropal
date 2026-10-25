import {api, apiData, apiWithHeaders} from "@/lib/api/client";
import {PostAuthor, PostCommentSchema, PostDetailSchema, PostSchema} from "@/app/(app)/discussions/_types/post.schema";
import {ProductPriceSchema, ProductSchema, ProductStockSchema} from "@/app/(app)/products/_types/product.schema";
import {TransactionSchema, TransactionType} from "@/app/(app)/orders/_types/order.schema";

// Calls to the Agropal backend, mapped onto the shapes the pages already use.

// ── Wire types (see backend/docs/openapi.yaml) ──
export interface UserDto {
    user_id: string;
    name: string;
    email: string;
    username: string;
    avatar: string;
    role: string;
}

export interface FarmDto {
    farm_id: string;
    user_id: string;
    name: string;
}

export interface LocationDto {
    location_id: string;
    longitude: number;
    latitude: number;
}

// Stock coming in is "income", going out is "expense", the same words orders use
export type StockType = "INCOME" | "EXPENSE";

export interface ProductDetailDto {
    product: {product_id: string, farm_id: string, name: string, description: string, image: string, unit: string};
    stocks: {stock_id: string, product_id: string, date_added: string, type: StockType, stock: number}[] | null;
    prices: {price_id: string, product_id: string, date_added: string, price: number, currency: string}[] | null;
}

export interface TransactionDto {
    transaction_id: string;
    product_id: string;
    date: string;
    type: TransactionType;
    amount: number;
    currency: string;
    description: string;
}

// ── User ──
export const getMe = () => api<UserDto>("/user/me");
export const updateMe = (patch: {name?: string, avatar?: string}) => api("/user/me", {method: "PATCH", body: patch});

// Ends every session, so the caller signs out afterwards
export const changePassword = (oldPassword: string, newPassword: string) =>
    api("/user/me/password", {method: "PATCH", body: {old_password: oldPassword, new_password: newPassword}});

// Closes the account and ends its sessions
export const deleteAccount = () => api("/user/me", {method: "DELETE"});

// ── Farm ──
export const listFarms = () => api<FarmDto[] | null>("/farm").then((farms) => farms ?? []);
// The create endpoints answer with the new record's id (or ids)
export const createFarm = (name: string) =>
    apiData<{farm_id: string}>("/farm", {method: "POST", body: {name}}).then((created) => created.farm_id);
// Takes the farm's products, prices, stock, locations and orders with it
export const deleteFarm = (farmID: string) => api(`/farm/${farmID}`, {method: "DELETE"});
export const updateFarm = (farmID: string, name: string) => api(`/farm/${farmID}`, {method: "PATCH", body: {name}});

// ── Farm location ──
export const listLocations = (farmID: string) => api<LocationDto[] | null>(`/farm/${farmID}/location`).then((l) => l ?? []);
export const createLocation = (farmID: string, latitude: number, longitude: number) =>
    apiData<{location_id: string}>(`/farm/${farmID}/location`, {method: "POST", body: {latitude, longitude}})
        .then((created) => created.location_id);
export const updateLocation = (farmID: string, locationID: string, latitude: number, longitude: number) =>
    api(`/farm/${farmID}/location/${locationID}`, {method: "PATCH", body: {latitude, longitude}});
export const deleteLocation = (farmID: string, locationID: string) =>
    api(`/farm/${farmID}/location/${locationID}`, {method: "DELETE"});

// ── Products ──
function toProduct(dto: ProductDetailDto): ProductSchema {
    const {product} = dto;
    const stocks: ProductStockSchema[] = (dto.stocks ?? []).map((s) => ({
        ID: s.stock_id,
        ProductID: s.product_id,
        DateAdded: new Date(s.date_added),
        Type: s.type,
        Stock: s.stock,
    }));
    const prices: ProductPriceSchema[] = (dto.prices ?? []).map((p) => ({
        ID: p.price_id,
        ProductID: p.product_id,
        DateAdded: new Date(p.date_added),
        Price: p.price,
        Currency: p.currency,
    }));
    const byDate = <T extends {DateAdded: Date}>(a: T, b: T) => a.DateAdded.getTime() - b.DateAdded.getTime();

    return {
        ID: product.product_id,
        Name: product.name,
        Description: product.description,
        Image: product.image,
        Unit: product.unit,
        FarmID: product.farm_id,
        // Oldest → newest, so the last entry is the latest
        Stocks: stocks.sort(byDate),
        Prices: prices.sort(byDate),
    };
}

export const listProducts = (farmID: string) =>
    api<ProductDetailDto[] | null>(`/farm/${farmID}/product`).then((list) => (list ?? []).map(toProduct));

export const getProductByID = (farmID: string, productID: string) =>
    api<ProductDetailDto>(`/farm/${farmID}/product/${productID}`).then(toProduct);

// image is a URL; "" for none
export const createProduct = (farmID: string, input: {name: string, description: string, image: string, unit: string}) =>
    apiData<{product_id: string}>(`/farm/${farmID}/product`, {method: "POST", body: input}).then((created) => created.product_id);

export const updateProduct = (farmID: string, productID: string, patch: {name?: string, description?: string, image?: string, unit?: string}) =>
    api(`/farm/${farmID}/product/${productID}`, {method: "PATCH", body: patch});

// Also removes the product's stock, prices, and orders (the database cascades)
export const deleteProduct = (farmID: string, productID: string) =>
    api(`/farm/${farmID}/product/${productID}`, {method: "DELETE"});

export const addPrice = (farmID: string, productID: string, price: number, currency: string, dateAdded = new Date()) =>
    apiData<{price_id: string}>(`/farm/${farmID}/product/${productID}/price`, {method: "POST", body: {price, currency, date_added: dateAdded.toISOString()}})
        .then((created) => created.price_id);

export const deletePrice = (farmID: string, productID: string, priceID: string) =>
    api(`/farm/${farmID}/product/${productID}/price/${priceID}`, {method: "DELETE"});

export const addStock = (farmID: string, productID: string, stock: number, type: StockType, dateAdded = new Date()) =>
    apiData<{stock_id: string}>(`/farm/${farmID}/product/${productID}/stock`, {method: "POST", body: {stock, type, date_added: dateAdded.toISOString()}})
        .then((created) => created.stock_id);

export const deleteStock = (farmID: string, productID: string, stockID: string) =>
    api(`/farm/${farmID}/product/${productID}/stock/${stockID}`, {method: "DELETE"});

// ── Orders (the backend calls them transactions) ──
function toTransaction(dto: TransactionDto): TransactionSchema {
    return {
        ID: dto.transaction_id,
        ProductID: dto.product_id,
        Date: new Date(dto.date),
        Type: dto.type,
        Amount: dto.amount,
        Currency: dto.currency,
        Description: dto.description,
    };
}

export const listTransactions = (farmID: string) =>
    api<TransactionDto[] | null>(`/farm/${farmID}/transaction`)
        .then((list) => (list ?? []).map(toTransaction).sort((a, b) => b.Date.getTime() - a.Date.getTime()));

export interface TransactionInput {
    date: Date;
    type: TransactionType;
    amount: number;
    currency: string;
    description: string;
}

export const createTransaction = (farmID: string, productID: string, input: TransactionInput) =>
    apiData<{transaction_ids: string[]}>(`/farm/${farmID}/product/${productID}/transaction`, {
        method: "POST",
        body: {transactions: [{...input, date: input.date.toISOString()}]},
    }).then((created) => created.transaction_ids[0]);

// Replaces every field: the backend validates the update like a create. The product can't change.
export const updateTransaction = (farmID: string, productID: string, transactionID: string, input: TransactionInput) =>
    api(`/farm/${farmID}/product/${productID}/transaction/${transactionID}`, {
        method: "PATCH",
        body: {...input, date: input.date.toISOString()},
    });

// One product's orders, newest first
export const listProductTransactions = (farmID: string, productID: string) =>
    api<TransactionDto[] | null>(`/farm/${farmID}/product/${productID}/transaction`)
        .then((list) => (list ?? []).map(toTransaction).sort((a, b) => b.Date.getTime() - a.Date.getTime()));

// Every order on every farm the user owns
export const deleteAllTransactions = () => api("/transaction/all", {method: "DELETE"});

// Any of the user's transactions, across products; the backend checks ownership of each
export const deleteTransactions = (transactionIDs: string[]) =>
    api("/transaction", {method: "DELETE", body: {transaction_ids: transactionIDs}});

// ── Discussions ──
export interface PostDto {
    post_id: string;
    user_id: string;
    title: string;
    content: string;
    slug: string;
    like_count: number;
    dislike_count: number;
    view_count: number;
    comment_count: number;
    author_name: string;
    viewer_reaction?: "LIKE" | "DISLIKE";
    created_at: string;
    updated_at: string | null;
}

export interface CommentDto {
    comment_id: string;
    post_id: string;
    user_id: string;
    parent_id: string | null;
    content: string;
    reply_count: number;
    like_count: number;
    dislike_count: number;
    view_count: number;
    author_name: string;
    viewer_reaction?: "LIKE" | "DISLIKE";
    created_at: string;
    updated_at: string | null;
}

// The backend stamps updated_at when a post or comment is created, so only a later change counts as an edit
function editedAt(dto: {created_at: string, updated_at: string | null}): Date | undefined {
    if (!dto.updated_at) return undefined;
    const updated = new Date(dto.updated_at);
    return updated.getTime() - new Date(dto.created_at).getTime() > 60_000 ? updated : undefined;
}

function toPost(dto: PostDto): PostSchema {
    return {
        ID: dto.post_id,
        UserID: dto.user_id,
        Title: dto.title,
        Content: dto.content,
        Slug: dto.slug,
        LikeCount: dto.like_count,
        DislikeCount: dto.dislike_count,
        ViewCount: dto.view_count,
        CommentCount: dto.comment_count,
        AuthorName: dto.author_name,
        ViewerReaction: dto.viewer_reaction,
        CreatedAt: new Date(dto.created_at),
        UpdatedAt: editedAt(dto),
    };
}

function toComment(dto: CommentDto): PostCommentSchema {
    return {
        ID: dto.comment_id,
        PostID: dto.post_id,
        UserID: dto.user_id,
        ParentID: dto.parent_id ?? undefined,
        Content: dto.content,
        ReplyCount: dto.reply_count,
        LikeCount: dto.like_count,
        DislikeCount: dto.dislike_count,
        ViewCount: dto.view_count,
        AuthorName: dto.author_name,
        ViewerReaction: dto.viewer_reaction,
        CreatedAt: new Date(dto.created_at),
        UpdatedAt: editedAt(dto),
    };
}

// The values the backend accepts for sort_column; anything else is rejected with a 400
export const POST_SORTS = {
    "newest": {column: "created_at", dir: "desc"},
    "oldest": {column: "created_at", dir: "asc"},
    "most-liked": {column: "like_count", dir: "desc"},
    "most-commented": {column: "comment_count", dir: "desc"},
    "most-viewed": {column: "view_count", dir: "desc"},
} as const;

export type PostSortKey = keyof typeof POST_SORTS;

export interface PostQuery {
    sort: PostSortKey;
    authorIDs: string[];
    search: string;
    page: number;
    pageSize: number;
}

export interface PostPage {
    posts: PostSchema[];
    total: number;
}

export async function listPosts({sort, authorIDs, search, page, pageSize}: PostQuery): Promise<PostPage> {
    const {column, dir} = POST_SORTS[sort];
    const params = new URLSearchParams({
        sort_column: column,
        sort_dir: dir,
        limit: String(pageSize),
        offset: String((page - 1) * pageSize),
    });
    authorIDs.forEach((id) => params.append("user_id", id));
    if (search.trim()) params.set("q", search.trim());

    const {data, headers} = await apiWithHeaders<PostDto[] | null>(`/post?${params}`);
    const posts = (data ?? []).map(toPost);
    // The total counts every match, not just this page
    return {posts, total: Number(headers.get("X-Total-Count") ?? posts.length)};
}

export const listPostAuthors = (): Promise<PostAuthor[]> =>
    api<{user_id: string, name: string}[] | null>("/post/authors")
        .then((authors) => (authors ?? []).map((a) => ({ID: a.user_id, Name: a.name})));

// Detail is by id, but pages are addressed by slug. The slug filter matches substrings, so the
// exact match is picked here; null when there is none.
export async function getPostBySlug(slug: string): Promise<PostDetailSchema | null> {
    const matches = await api<PostDto[] | null>(`/post?${new URLSearchParams({slug})}`);
    const match = (matches ?? []).find((post) => post.slug === slug);
    if (!match) return null;

    const detail = await api<{post: PostDto, comments: CommentDto[] | null}>(`/post/${match.post_id}`);
    return {
        Post: toPost(detail.post),
        Comments: (detail.comments ?? []).map(toComment).sort((a, b) => a.CreatedAt.getTime() - b.CreatedAt.getTime()),
    };
}

// slug is the post's address in the app
export const createPost = (title: string, content: string) =>
    apiData<{post_id: string, slug: string}>("/post", {method: "POST", body: {title, content}});

// parentID makes it a reply to another comment on the same post
export const createComment = (postID: string, content: string, parentID?: string) =>
    apiData<{comment_id: string}>(`/post/${postID}/comment`, {method: "POST", body: {content, ...(parentID && {parent_id: parentID})}})
        .then((created) => created.comment_id);

// A user has one like or dislike per post or comment: sending the other one switches it
export type EngagementType = "LIKE" | "DISLIKE" | "VIEW";

export const addEngagement = (target: "post" | "comment", id: string, type: EngagementType) =>
    api(`/${target}/${id}/engagement`, {method: "POST", body: {type}});

// Takes back the user's like or dislike
export const removeEngagement = (target: "post" | "comment", id: string) =>
    api(`/${target}/${id}/engagement`, {method: "DELETE"});
