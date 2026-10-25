import type {CommentDto, PostDto, ProductDetailDto, TransactionDto} from "@/lib/api/resources";
import {DEMO_FARM_ID} from "@/lib/demo/base";

// Sample data for the public demo: a small mixed vegetable farm after a season of use. Dates are
// relative to today so the dashboard's current-year chart always has months to show. Everything is
// deterministic, so a reload shows the same farm.

import {DEMO_USER} from "@/lib/demo/identity";

export {DEMO_FARM, DEMO_LOCATION, DEMO_LOCATION_NAME, DEMO_USER} from "@/lib/demo/identity";

// ── A tiny seeded generator, so the numbers don't change between visits ──
function seeded(seed: number) {
    let a = seed;
    return () => {
        a = (a + 0x6d2b79f5) | 0;
        let t = Math.imul(a ^ (a >>> 15), 1 | a);
        t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
        return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    };
}

const DAY = 86_400_000;
const NOW = Date.now();
const daysAgo = (days: number, hour = 9) => {
    const d = new Date(NOW - days * DAY);
    d.setUTCHours(hour, 0, 0, 0);
    return d.toISOString();
};

interface ProductSpec {
    key: string;
    name: string;
    description: string;
    unit: string;
    // Price per unit over the year, oldest first; the last is the current price
    prices: number[];
    // Typical size of one harvest, in the product's unit
    harvest: number;
    // Stores well, so it sells all year; otherwise it sells in season and for a month after
    keeps: boolean;
    // Which months (0-11) the crop comes in; sales run all year
    season: number[];
    salesNote: string;
}

const PRODUCTS: ProductSpec[] = [
    {key: "tomatoes", name: "Heirloom tomatoes", description: "Brandywine, Cherokee Purple and Green Zebra, picked ripe and sold the same week.", unit: "kg", prices: [5.5, 6, 6.5], harvest: 70, keeps: false, season: [5, 6, 7, 8, 9], salesNote: "Saturday farmers market"},
    {key: "corn", name: "Sweet corn", description: "Bicolor sweet corn, picked in the morning and sold by weight at the roadside stand.", unit: "kg", prices: [3.2, 3.5], harvest: 140, keeps: false, season: [6, 7, 8], salesNote: "Roadside stand"},
    {key: "eggs", name: "Pasture-raised eggs", description: "Mixed-breed hens on rotating pasture. Sold by the dozen.", unit: "dozen", prices: [5, 5.5, 6], harvest: 16, keeps: true, season: [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11], salesNote: "Weekly subscription boxes"},
    {key: "honey", name: "Wildflower honey", description: "Raw honey from the hives along the creek, 340 g jars.", unit: "jar", prices: [9, 10, 11], harvest: 70, keeps: true, season: [7, 8, 9], salesNote: "Farm shop"},
    {key: "lettuce", name: "Salad lettuce", description: "Mixed red and green leaf lettuce, cut to order for restaurants.", unit: "kg", prices: [7, 7.5], harvest: 45, keeps: false, season: [3, 4, 5, 8, 9, 10], salesNote: "Restaurant delivery"},
    {key: "carrots", name: "Rainbow carrots", description: "Orange, purple and yellow carrots, bunched with their tops.", unit: "kg", prices: [3.8, 4, 4.2], harvest: 90, keeps: true, season: [5, 6, 7, 8, 9, 10], salesNote: "Saturday farmers market"},
    {key: "strawberries", name: "June-bearing strawberries", description: "Sweet, small-berried strawberries. A short season, and they sell out fast.", unit: "kg", prices: [9, 9.5], harvest: 60, keeps: false, season: [4, 5], salesNote: "Pick-your-own weekends"},
    {key: "wheat", name: "Hard red winter wheat", description: "Stone-milling wheat, sold in 25 kg sacks to the village bakery.", unit: "sack", prices: [13.5, 14.5], harvest: 40, keeps: true, season: [6, 7], salesNote: "Bakery contract"},
    {key: "peppers", name: "Sweet bell peppers", description: "Red and yellow bell peppers from the high tunnel.", unit: "kg", prices: [6, 6.5, 6.5], harvest: 70, keeps: false, season: [6, 7, 8, 9], salesNote: "CSA boxes"},
    {key: "onions", name: "Red onions", description: "Cured red onions that keep through winter.", unit: "kg", prices: [2.4, 2.6], harvest: 170, keeps: true, season: [7, 8, 9], salesNote: "Wholesale to the co-op"},
];

const EXPENSES = [
    {note: "Seed order", min: 60, max: 180},
    {note: "Compost delivery", min: 90, max: 260},
    {note: "Seasonal labour", min: 150, max: 420},
    {note: "Crates and packaging", min: 40, max: 110},
    {note: "Irrigation repairs", min: 35, max: 140},
    {note: "Market stall fee", min: 25, max: 45},
];

const productId = (key: string) => `demo-product-${key}`;

function build() {
    const rand = seeded(20240611);
    const products: ProductDetailDto[] = [];
    const transactions: TransactionDto[] = [];
    let txCount = 0;
    let stockCount = 0;
    const thisYear = new Date(NOW).getUTCFullYear();

    for (const spec of PRODUCTS) {
        const id = productId(spec.key);
        // Sales roughly match what the harvest brings in, spread over the weeks the product sells
        const saleMonths = spec.keeps ? 12 : spec.season.length + 1;
        const sale = spec.harvest * 0.75 * (spec.season.length * 4.3 * 0.75) / (saleMonths * 4.3 * 0.8);
        const sells = (month: number) => spec.keeps || spec.season.includes(month) || spec.season.includes((month + 11) % 12);
        const stocks: NonNullable<ProductDetailDto["stocks"]> = [];
        const prices: NonNullable<ProductDetailDto["prices"]> = [];

        // Prices step up across the year
        spec.prices.forEach((price, i) => {
            prices.push({
                price_id: `demo-price-${spec.key}-${i}`,
                product_id: id,
                date_added: daysAgo(330 - i * Math.floor(300 / spec.prices.length)),
                price,
                currency: "USD",
            });
        });

        // Walk back through roughly 11 months, a week at a time
        for (let week = 48; week >= 0; week--) {
            const date = new Date(NOW - week * 7 * DAY);
            const month = date.getUTCMonth();
            const inSeason = spec.season.includes(month);
            const price = [...prices].reverse().find((p) => new Date(p.date_added) <= date)?.price ?? spec.prices[0];

            if (inSeason && (spec.key === "eggs" || rand() < 0.75)) {
                const amount = Math.max(1, Math.round(spec.harvest * (0.55 + rand() * 0.9)));
                stocks.push({stock_id: `demo-stock-${stockCount++}`, product_id: id, date_added: daysAgo(week * 7 + 1, 7), type: "INCOME", stock: amount});
            }

            // Sales happen most weeks, and a little spoils or is given away now and then
            if (sells(month) && rand() < 0.8) {
                const amount = Math.max(1, Math.round(sale * (0.5 + rand())));
                stocks.push({stock_id: `demo-stock-${stockCount++}`, product_id: id, date_added: daysAgo(week * 7, 13), type: "EXPENSE", stock: amount});
                transactions.push({
                    transaction_id: `demo-order-${txCount++}`,
                    product_id: id,
                    date: daysAgo(week * 7, 13),
                    type: "INCOME",
                    amount: Math.round(amount * price * 100) / 100,
                    currency: "USD",
                    description: `${spec.salesNote}: ${amount} ${spec.unit}`,
                });
            }
            if (sells(month) && rand() < 0.07) {
                stocks.push({stock_id: `demo-stock-${stockCount++}`, product_id: id, date_added: daysAgo(week * 7 + 2, 16), type: "EXPENSE", stock: Math.max(1, Math.round(sale * 0.2))});
            }
            if (rand() < 0.1) {
                const expense = EXPENSES[Math.floor(rand() * EXPENSES.length)];
                transactions.push({
                    transaction_id: `demo-order-${txCount++}`,
                    product_id: id,
                    date: daysAgo(week * 7 + 3, 11),
                    type: "EXPENSE",
                    amount: Math.round(expense.min + rand() * (expense.max - expense.min)),
                    currency: "USD",
                    description: expense.note,
                });
            }
        }

        // Guarantee this year has movement for the chart even early in January
        if (!stocks.some((s) => new Date(s.date_added).getUTCFullYear() === thisYear)) {
            stocks.push({stock_id: `demo-stock-${stockCount++}`, product_id: id, date_added: daysAgo(1, 8), type: "INCOME", stock: spec.harvest});
        }

        products.push({
            product: {product_id: id, farm_id: DEMO_FARM_ID, name: spec.name, description: spec.description, image: `/demo/${spec.key}.jpg`, unit: spec.unit},
            stocks,
            prices,
        });
    }

    transactions.sort((a, b) => b.date.localeCompare(a.date));
    return {products, transactions};
}

const built = build();
export const DEMO_PRODUCTS = built.products;
export const DEMO_TRANSACTIONS = built.transactions;

// ── Discussions ──
export const DEMO_AUTHORS = [
    {user_id: DEMO_USER.user_id, name: DEMO_USER.name},
    {user_id: "demo-user-tomas", name: "Tomás Herrera"},
    {user_id: "demo-user-priya", name: "Priya Nair"},
    {user_id: "demo-user-daniel", name: "Daniel Okafor"},
    {user_id: "demo-user-hannah", name: "Hannah Brooks"},
    {user_id: "demo-user-lars", name: "Lars Eriksen"},
];
const author = (name: string) => DEMO_AUTHORS.find((a) => a.name.startsWith(name))!;

interface PostSpec {
    slug: string;
    by: string;
    title: string;
    content: string;
    daysAgo: number;
    likes: number;
    dislikes?: number;
    views: number;
    comments: {id: string, by: string, text: string, hoursAfter: number, likes: number, parent?: string}[];
}

const POSTS: PostSpec[] = [
    {
        slug: "hornworms-on-my-tomatoes", by: "Tomás", daysAgo: 4, likes: 31, views: 412,
        title: "Hornworms are eating my tomatoes overnight. What actually works?",
        content: "I lost half a row of Brandywines to hornworms in three days. Hand-picking at dusk helps, but I can't keep up on two hundred plants.\n\nHas anyone had luck with Bt, or with encouraging wasps? I'd rather not spray anything near the market stand.",
        comments: [
            {id: "a", by: "Priya", text: "Bt (Bacillus thuringiensis) worked for us, sprayed in the evening so the bees are not around. It only affects caterpillars. Repeat after rain.", hoursAfter: 3, likes: 14},
            {id: "b", by: "Tomás", text: "Does it matter how early you catch them? I only noticed once the damage was big.", hoursAfter: 5, likes: 2, parent: "a"},
            {id: "c", by: "Priya", text: "Yes, the small ones die within a day or two of eating it. The big ones stop eating but take longer to go.", hoursAfter: 6, likes: 5, parent: "b"},
            {id: "d", by: "Hannah", text: "Plant dill and borage nearby. We get braconid wasps every year now, and you will see the caterpillars covered in little white cocoons. Leave those alone.", hoursAfter: 9, likes: 11},
            {id: "e", by: "Maya", text: "Also check the undersides of the leaves for frass. The dark droppings give them away before the damage does.", hoursAfter: 20, likes: 6},
        ],
    },
    {
        slug: "farmers-market-pricing-for-eggs", by: "Maya", daysAgo: 9, likes: 24, views: 301,
        title: "Are you raising your egg price this year? Feed is up again",
        content: "I've held at $5.50 a dozen since spring but feed went up again. A couple of stalls near me moved to $7 and they're still selling out.\n\nHow do you decide when to raise prices without losing regulars?",
        comments: [
            {id: "a", by: "Hannah", text: "We sent a short note in the subscription box explaining the feed cost and moved to $6.50. Lost one customer out of forty.", hoursAfter: 2, likes: 9},
            {id: "b", by: "Daniel", text: "Pasture-raised eggs are priced against the supermarket's organic ones, not against the cheap ones. Check what those are this week.", hoursAfter: 7, likes: 7},
            {id: "c", by: "Lars", text: "Keep a log of the cost per dozen. Once you know it, the price is a lot less emotional.", hoursAfter: 12, likes: 4, parent: "b"},
        ],
    },
    {
        slug: "when-did-you-plant-out-after-last-frost", by: "Priya", daysAgo: 21, likes: 18, views: 266,
        title: "How long after the last frost do you wait before planting out peppers?",
        content: "Our average last frost is mid-May but the soil stays cold for another two weeks. I lost a flat of pepper seedlings last year by planting out too eagerly.\n\nDo you wait for a soil temperature, or for the calendar?",
        comments: [
            {id: "a", by: "Daniel", text: "Soil temperature, 18 °C at 10 cm for three days in a row. A cheap probe thermometer pays for itself.", hoursAfter: 4, likes: 10},
            {id: "b", by: "Maya", text: "The weather page here shows soil temperature for your own field, which is what I check each morning in May.", hoursAfter: 8, likes: 3},
        ],
    },
    {
        slug: "cover-crop-mix-after-garlic", by: "Lars", daysAgo: 34, likes: 15, views: 198,
        title: "Cover crop mix after garlic harvest?",
        content: "The garlic comes out in mid-July and the bed sits empty until spring. I'm considering buckwheat followed by winter rye, but I'm not sure the rye will terminate cleanly with a roller-crimper.\n\nWhat has worked for you?",
        comments: [
            {id: "a", by: "Tomás", text: "Buckwheat flowers within a month and the bees love it. Cut it before it sets seed or you will have it forever.", hoursAfter: 6, likes: 8},
            {id: "b", by: "Hannah", text: "Rye terminates best at anthesis. If you roll it earlier, it comes back.", hoursAfter: 26, likes: 6},
        ],
    },
    {
        slug: "selling-honey-do-i-need-a-licence", by: "Daniel", daysAgo: 47, likes: 12, views: 187,
        title: "Selling honey at the market: what paperwork did you need?",
        content: "I have six hives and finally have more honey than the family can eat. Is a cottage food licence enough, or do I need a separate permit for honey?",
        comments: [
            {id: "a", by: "Maya", text: "Honey counts as a low-risk food in most states, but the label rules are strict. Check what your state department of agriculture requires for net weight and origin.", hoursAfter: 5, likes: 7},
        ],
    },
    {
        slug: "drip-irrigation-zones-for-small-plots", by: "Hannah", daysAgo: 63, likes: 22, views: 340,
        title: "Drip irrigation zones for a half-acre market garden",
        content: "I'm replacing the hand-watering routine with drip lines. Is it worth splitting the plot into separate zones by crop, or is one big loop fine?\n\nThe lettuce and the peppers want very different amounts of water.",
        comments: [
            {id: "a", by: "Lars", text: "Zones by crop, absolutely. One valve per bed group and a timer for each. The savings in water pay for the valves in a season.", hoursAfter: 3, likes: 12},
            {id: "b", by: "Priya", text: "Put a pressure regulator at the head of each zone, or the far end of the line will starve.", hoursAfter: 10, likes: 9},
            {id: "c", by: "Hannah", text: "Thanks both, ordering regulators today.", hoursAfter: 11, likes: 1, parent: "b"},
        ],
    },
    {
        slug: "storing-onions-through-winter", by: "Tomás", daysAgo: 80, likes: 9, views: 144,
        title: "Storing red onions through winter: curing tips",
        content: "My red onions sprouted by January last year. This year I'm curing them for two weeks in the barn before storing. Anything else I should do?",
        comments: [
            {id: "a", by: "Daniel", text: "Cure until the necks are completely dry, then store cold and dry, around 2 °C. Red onions are not the best keepers, so sell those first.", hoursAfter: 8, likes: 5},
        ],
    },
    {
        slug: "strawberry-bed-renovation", by: "Maya", daysAgo: 96, likes: 14, views: 176,
        title: "Strawberry bed renovation after the June harvest",
        content: "I mow the leaves right after harvest, thin the plants to six inches apart, and fertilise lightly. The beds are producing less each year, though, so I wonder if I should replant every three years instead.",
        comments: [
            {id: "a", by: "Hannah", text: "Three to four years is about right for June-bearers. Take runners from your best plants and start a fresh bed each year in rotation.", hoursAfter: 5, likes: 8},
            {id: "b", by: "Lars", text: "Straw mulch in winter made a visible difference to our yield.", hoursAfter: 14, likes: 3},
        ],
    },
    {
        slug: "csa-box-sizes-that-actually-sell", by: "Priya", daysAgo: 130, likes: 27, views: 389,
        title: "CSA box sizes that actually sell",
        content: "We started with one large box at $35 and found it was too much for singles and couples. Splitting it into a small ($22) and a regular ($35) made signups jump.\n\nWhat sizes and prices work for your area?",
        comments: [
            {id: "a", by: "Maya", text: "We do a half-share every other week as well, and it brought in the people who were worried about waste.", hoursAfter: 6, likes: 10},
            {id: "b", by: "Daniel", text: "Allow a swap list. People happily trade away the vegetables they don't like and stay subscribed longer.", hoursAfter: 15, likes: 13},
        ],
    },
    {
        slug: "winter-wheat-seeding-dates", by: "Lars", daysAgo: 160, likes: 8, views: 121,
        title: "Seeding dates for hard red winter wheat in zone 5",
        content: "Is late September too early for winter wheat in central Iowa? I want good tillering before winter but not so much growth that the Hessian fly gets it.",
        comments: [
            {id: "a", by: "Tomás", text: "The fly-free date for your county is the one to follow. Your extension office publishes it each year.", hoursAfter: 9, likes: 6},
        ],
    },
];

const hours = (days: number, extra = 0) => new Date(NOW - days * DAY + extra * 3_600_000).toISOString();

function buildPosts() {
    const posts: PostDto[] = [];
    const comments: CommentDto[] = [];

    for (const spec of POSTS) {
        const owner = author(spec.by);
        const postId = `demo-post-${spec.slug}`;
        const created = hours(spec.daysAgo);

        posts.push({
            post_id: postId,
            user_id: owner.user_id,
            title: spec.title,
            content: spec.content,
            slug: spec.slug,
            like_count: spec.likes,
            dislike_count: spec.dislikes ?? 0,
            view_count: spec.views,
            comment_count: spec.comments.length,
            author_name: owner.name,
            created_at: created,
            updated_at: created,
        });

        for (const c of spec.comments) {
            const writer = author(c.by);
            const stamp = hours(spec.daysAgo, c.hoursAfter);
            comments.push({
                comment_id: `demo-comment-${spec.slug}-${c.id}`,
                post_id: postId,
                user_id: writer.user_id,
                parent_id: c.parent ? `demo-comment-${spec.slug}-${c.parent}` : null,
                content: c.text,
                reply_count: spec.comments.filter((x) => x.parent === c.id).length,
                like_count: c.likes,
                dislike_count: 0,
                view_count: c.likes * 6,
                author_name: writer.name,
                created_at: stamp,
                updated_at: stamp,
            });
        }
    }
    return {posts, comments};
}

const builtPosts = buildPosts();
export const DEMO_POSTS = builtPosts.posts;
export const DEMO_COMMENTS = builtPosts.comments;
