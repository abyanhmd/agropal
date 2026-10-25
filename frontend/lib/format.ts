// Display formatters shared across pages

export const numberFormat = new Intl.NumberFormat("en-US");
// UTC to match how the mock dates are created, so server and browser render the same text
export const shortDate = new Intl.DateTimeFormat("en-US", {month: "short", day: "numeric", timeZone: "UTC"});

const currencyFormats = new Map<string, Intl.NumberFormat>();

export function formatPrice(amount: number, currency: string): string {
    const code = currency.toUpperCase();
    let format = currencyFormats.get(code);

    if (!format) {
        try {
            format = new Intl.NumberFormat("en-US", {style: "currency", currency: code, currencyDisplay: "narrowSymbol"});
        } catch {
            // Unknown currency code: show the number with the raw code
            return `${amount.toLocaleString("en-US")} ${currency}`;
        }
        currencyFormats.set(code, format);
    }

    return format.format(amount);
}
