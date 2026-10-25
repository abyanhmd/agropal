import {ProductSchema, ProductStockSchema} from "@/app/(app)/products/_types/product.schema";

// All dashboard numbers for one unit. Units are never mixed: adding kg to pcs is meaningless.

export interface MonthlyMovement {
    month: string;   // "January"
    stockIn: number;
    stockOut: number;
}

export interface ProductNet {
    id: string;
    name: string;
    net: number;
}

export interface Activity {
    product: {id: string, name: string};
    stock: ProductStockSchema;
}

export interface DashboardSummary {
    productCount: number;
    stockIn: number;
    stockOut: number;
    net: number;
    monthly: MonthlyMovement[];
    mostStocked: ProductNet[];
    mostDepleted: ProductNet[];
    recent: Activity[];
}

export function unitsOf(products: ProductSchema[]): string[] {
    return [...new Set(products.map((p) => p.Unit).filter(Boolean))].sort();
}

// The unit most products use is the default (alphabetically among ties), and is left out of the URL:
// "/dashboard" vs "/dashboard?unit=pcs"
export function defaultUnit(products: ProductSchema[]): string {
    const counts = new Map<string, number>();
    for (const product of products) if (product.Unit) counts.set(product.Unit, (counts.get(product.Unit) ?? 0) + 1);
    return unitsOf(products).sort((a, b) => (counts.get(b) ?? 0) - (counts.get(a) ?? 0))[0] ?? "";
}

// The year the monthly chart covers
export const CHART_YEAR = new Date().getUTCFullYear();

const MONTHS = [
    "January", "February", "March", "April", "May", "June",
    "July", "August", "September", "October", "November", "December",
];

export function summarize(products: ProductSchema[], unit: string): DashboardSummary {
    const scoped = products.filter((p) => p.Unit === unit);
    const monthly: MonthlyMovement[] = MONTHS.map((month) => ({month, stockIn: 0, stockOut: 0}));
    const nets: ProductNet[] = [];
    const recent: Activity[] = [];
    let stockIn = 0;
    let stockOut = 0;

    for (const product of scoped) {
        let net = 0;
        for (const stock of product.Stocks ?? []) {
            // Dates are UTC, so bucket by UTC month. Totals count everything; the monthly chart is this year.
            const added = new Date(stock.DateAdded);
            const bucket = added.getUTCFullYear() === CHART_YEAR ? monthly[added.getUTCMonth()] : null;
            if (stock.Type === "INCOME") {
                stockIn += stock.Stock;
                if (bucket) bucket.stockIn += stock.Stock;
                net += stock.Stock;
            } else {
                stockOut += stock.Stock;
                if (bucket) bucket.stockOut += stock.Stock;
                net -= stock.Stock;
            }
            recent.push({product: {id: product.ID, name: product.Name}, stock});
        }
        nets.push({id: product.ID, name: product.Name, net});
    }

    recent.sort((a, b) => new Date(b.stock.DateAdded).getTime() - new Date(a.stock.DateAdded).getTime());

    return {
        productCount: scoped.length,
        stockIn,
        stockOut,
        net: stockIn - stockOut,
        monthly,
        mostStocked: nets.filter((n) => n.net > 0).sort((a, b) => b.net - a.net).slice(0, 5),
        mostDepleted: nets.filter((n) => n.net < 0).sort((a, b) => a.net - b.net).slice(0, 5),
        recent: recent.slice(0, 8),
    };
}
