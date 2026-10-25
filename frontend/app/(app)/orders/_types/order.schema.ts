// Mirrors the backend Transaction entity; the frontend calls them orders
export type TransactionType = "INCOME" | "EXPENSE";

export interface TransactionSchema {
    ID: string,
    ProductID: string,
    Date: Date,
    Type: TransactionType,
    Amount: number,
    Currency: string,
    Description: string,
}

export type TransactionSchemaType = TransactionSchema;
