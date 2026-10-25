// Currencies offered in price and order forms (ISO 4217, which the backend validates)
export const CURRENCIES = ["USD", "GBP", "EUR", "IDR"].map((code) => ({value: code, label: code}));
