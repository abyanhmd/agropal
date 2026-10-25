import {
    DEMO_AUTHORS,
    DEMO_COMMENTS,
    DEMO_POSTS,
    DEMO_PRODUCTS,
    DEMO_TRANSACTIONS,
} from "@/lib/demo/data";
import {DEMO_FARM, DEMO_LOCATION, DEMO_USER} from "@/lib/demo/identity";
import {DEMO_FARM_ID} from "@/lib/demo/base";

// A stand-in for the backend while the demo is open. It answers the same requests the pages make,
// with the same shapes, from the sample data, so the real pages run on it unchanged. Nothing is
// ever written: every change is answered with a note that this is a demo.

export interface DemoResponse {
    status: number;
    data: unknown;
    headers?: Record<string, string>;
    message?: string;
}

const READ_ONLY = "This is a demo with sample data, so changes aren't saved. Create an account to start your own farm.";

const ok = (data: unknown, headers?: Record<string, string>): DemoResponse => ({status: 200, data, headers});
const notFound = (): DemoResponse => ({status: 404, data: null, message: "That doesn't exist in the demo."});

function listPosts(params: URLSearchParams): DemoResponse {
    const slug = params.get("slug");
    const query = params.get("q")?.trim().toLowerCase();
    const authors = params.getAll("user_id");

    let posts = DEMO_POSTS.filter((post) => {
        if (slug && !post.slug.includes(slug)) return false;
        if (authors.length > 0 && !authors.includes(post.user_id)) return false;
        if (query && !post.title.toLowerCase().includes(query) && !post.content.toLowerCase().includes(query)) return false;
        return true;
    });

    const column = (params.get("sort_column") ?? "created_at") as "created_at" | "like_count" | "comment_count" | "view_count";
    const direction = params.get("sort_dir") === "asc" ? 1 : -1;
    posts = [...posts].sort((a, b) => {
        const x = a[column];
        const y = b[column];
        return (x < y ? -1 : x > y ? 1 : 0) * direction;
    });

    const total = posts.length;
    const offset = Number(params.get("offset") ?? 0);
    const limit = Number(params.get("limit") ?? 20);
    return ok(posts.slice(offset, offset + limit), {"X-Total-Count": String(total)});
}

export function demoRequest(method: string, rawPath: string): DemoResponse {
    const [path, queryString = ""] = rawPath.split("?");
    const params = new URLSearchParams(queryString);

    if (method !== "GET") return {status: 403, data: null, message: READ_ONLY};

    if (path === "/user/me") return ok(DEMO_USER);
    if (path === "/farm") return ok([DEMO_FARM]);

    if (path === `/farm/${DEMO_FARM_ID}/location`) return ok([DEMO_LOCATION]);
    if (path === `/farm/${DEMO_FARM_ID}/product`) return ok(DEMO_PRODUCTS);
    if (path === `/farm/${DEMO_FARM_ID}/transaction`) return ok(DEMO_TRANSACTIONS);

    const product = path.match(new RegExp(`^/farm/${DEMO_FARM_ID}/product/([^/]+)$`));
    if (product) {
        const found = DEMO_PRODUCTS.find((p) => p.product.product_id === product[1]);
        return found ? ok(found) : notFound();
    }

    const productOrders = path.match(new RegExp(`^/farm/${DEMO_FARM_ID}/product/([^/]+)/transaction$`));
    if (productOrders) return ok(DEMO_TRANSACTIONS.filter((t) => t.product_id === productOrders[1]));

    if (path === "/post") return listPosts(params);
    if (path === "/post/authors") return ok(DEMO_AUTHORS);

    const post = path.match(/^\/post\/([^/]+)$/);
    if (post) {
        const found = DEMO_POSTS.find((p) => p.post_id === post[1]);
        if (!found) return notFound();
        return ok({post: found, comments: DEMO_COMMENTS.filter((c) => c.post_id === found.post_id)});
    }

    return notFound();
}
