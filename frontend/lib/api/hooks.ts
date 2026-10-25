'use client';

import {useAuth, useFarm} from "@/lib/auth/auth-provider";
import {forget, invalidate, QueryState, useQuery} from "@/lib/api/use-query";
import {
    addEngagement,
    addPrice,
    createComment,
    createPost,
    addStock,
    createProduct,
    createTransaction,
    deletePrice,
    deleteProduct,
    deleteStock,
    deleteAllTransactions,
    deleteTransactions,
    getPostBySlug,
    getProductByID,
    listPostAuthors,
    listPosts,
    listProductTransactions,
    listProducts,
    PostPage,
    PostQuery,
    removeEngagement,
    listTransactions,
    StockType,
    updateProduct,
    updateTransaction,
} from "@/lib/api/resources";
import {PostAuthor, PostDetailSchema} from "@/app/(app)/discussions/_types/post.schema";
import {ProductSchema} from "@/app/(app)/products/_types/product.schema";
import {TransactionSchema, TransactionType} from "@/app/(app)/orders/_types/order.schema";
import {entryTimestamp, fromDateInput} from "@/lib/dates";

// Data for the pages, scoped to the signed-in user's farm. Mutations refetch what they touched.

export function useProducts(): QueryState<ProductSchema[]> {
    const farm = useFarm()?.farm;
    return useQuery(farm ? `products:${farm.farm_id}` : null, () => listProducts(farm!.farm_id));
}

export function useProduct(id: string): QueryState<ProductSchema> {
    const farm = useFarm()?.farm;
    return useQuery(farm ? `products:${farm.farm_id}:${id}` : null, () => getProductByID(farm!.farm_id, id));
}

// One product's orders, from the endpoint for exactly that instead of filtering the farm's whole list
export function useProductTransactions(productID: string): QueryState<TransactionSchema[]> {
    const farm = useFarm()?.farm;
    return useQuery(farm ? `orders:${farm.farm_id}:product:${productID}` : null, () => listProductTransactions(farm!.farm_id, productID));
}

export function useTransactions(): QueryState<TransactionSchema[]> {
    const farm = useFarm()?.farm;
    return useQuery(farm ? `orders:${farm.farm_id}` : null, () => listTransactions(farm!.farm_id));
}

// ── Products ──

export interface NewProduct {
    name: string;
    description: string;
    image: string;
    unit: string;
    price: number;
    currency: string;
    stock: number;
}

export async function addProduct(farmID: string, input: NewProduct): Promise<string> {
    const productID = await createProduct(farmID, {name: input.name, description: input.description, image: input.image, unit: input.unit});

    try {
        if (input.price > 0) await addPrice(farmID, productID, input.price, input.currency);
        if (input.stock > 0) await addStock(farmID, productID, input.stock, "INCOME");
    } finally {
        // The product exists even if its price or stock didn't go through
        invalidate("products:");
    }
    return productID;
}

export interface ProductDetails {
    name: string;
    description: string;
    image: string;
}

export async function editProduct(farmID: string, productID: string, details: ProductDetails) {
    await updateProduct(farmID, productID, details);
    invalidate("products:");
}

// Navigate away afterwards: the product's page isn't refetched, since it would only show "not found"
export async function removeProduct(farmID: string, productID: string) {
    await deleteProduct(farmID, productID);
    // Its orders went with it
    forget("products:");
    forget("orders:");
}

export async function addProductPrice(farmID: string, productID: string, price: number, currency: string, day: string) {
    await addPrice(farmID, productID, price, currency, entryTimestamp(day));
    invalidate("products:");
}

export async function removeProductPrice(farmID: string, productID: string, priceID: string) {
    await deletePrice(farmID, productID, priceID);
    invalidate("products:");
}

export async function addProductStock(farmID: string, productID: string, stock: number, type: StockType, day: string) {
    await addStock(farmID, productID, stock, type, entryTimestamp(day));
    invalidate("products:");
}

export async function removeProductStock(farmID: string, productID: string, stockID: string) {
    await deleteStock(farmID, productID, stockID);
    invalidate("products:");
}

// ── Orders ──

export interface OrderDetails {
    type: TransactionType;
    amount: number;
    currency: string;
    date: string; // YYYY-MM-DD
    description: string;
}

export interface NewOrder extends OrderDetails {
    productID: string;
}

const toTransactionInput = (details: OrderDetails) => ({
    // UTC midnight, like the dates shown
    date: fromDateInput(details.date),
    type: details.type,
    amount: Math.round(details.amount * 100) / 100,
    currency: details.currency,
    description: details.description,
});

export async function addOrder(farmID: string, input: NewOrder): Promise<string> {
    const orderID = await createTransaction(farmID, input.productID, toTransactionInput(input));
    invalidate("orders:");
    return orderID;
}

export async function editOrder(farmID: string, order: TransactionSchema, details: OrderDetails) {
    await updateTransaction(farmID, order.ProductID, order.ID, toTransactionInput(details));
    invalidate("orders:");
}

// `leaving`: the caller navigates away (deleting from the order's own page), so skip the refetch
export async function removeOrders(orderIDs: string[], {leaving = false} = {}) {
    await deleteTransactions(orderIDs);
    if (leaving) forget("orders:");
    else invalidate("orders:");
}

export async function removeAllOrders() {
    await deleteAllTransactions();
    invalidate("orders:");
}

// ── Discussions ──
// Posts are shared between farms, so these keys aren't scoped to the farm

export function usePosts(query: PostQuery): QueryState<PostPage> {
    const {session} = useAuth();
    const {sort, authorIDs, search, page, pageSize} = query;
    const key = `posts:list:${sort}:${authorIDs.join(",")}:${search.trim()}:${page}:${pageSize}`;
    return useQuery(session.status === "ready" ? key : null, () => listPosts(query));
}

export function usePostAuthors(): QueryState<PostAuthor[]> {
    const {session} = useAuth();
    return useQuery(session.status === "ready" ? "posts:authors" : null, listPostAuthors);
}

export function usePost(slug: string): QueryState<PostDetailSchema | null> {
    const {session} = useAuth();
    return useQuery(session.status === "ready" ? `posts:detail:${slug}` : null, () => getPostBySlug(slug));
}

// Resolves to the new post's slug, its address in the app
export async function addPost(title: string, content: string): Promise<string> {
    const {slug} = await createPost(title, content);
    invalidate("posts:");
    return slug;
}

export async function addPostComment(postID: string, content: string, parentID?: string) {
    await createComment(postID, content, parentID);
    invalidate("posts:");
}

// null takes the reaction back; the counts refresh afterwards
export async function reactTo(target: "post" | "comment", id: string, type: "LIKE" | "DISLIKE" | null) {
    if (type) await addEngagement(target, id, type);
    else await removeEngagement(target, id);
    invalidate("posts:");
}

// A view is counted once per post per browser session
export function recordView(postID: string) {
    const key = `agropal.viewed.${postID}`;
    try {
        if (sessionStorage.getItem(key)) return;
        sessionStorage.setItem(key, "1");
    } catch {
        // Storage blocked: count it, at worst once per visit
    }
    addEngagement("post", postID, "VIEW").catch(() => {});
}
