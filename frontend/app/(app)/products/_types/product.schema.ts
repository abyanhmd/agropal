export interface ProductSchema {
    ID: string,
    Name: string,
    Description: string,
    Image: string,
    Unit: string,
    FarmID: string,
    Stocks?: ProductStockSchema[],
    Prices?: ProductPriceSchema[],
}

export type ProductSchemaType = ProductSchema;

export interface ProductStockSchema {
    ID: string,
    ProductID: string,
    DateAdded: Date,
    Type: string,
    Stock: number,
}

export type ProductStockSchemaType = ProductStockSchema;

export interface ProductPriceSchema {
    ID: string,
    ProductID: string,
    DateAdded: Date,
    Price: number,
    Currency: string,
}

export type ProductPriceSchemaType = ProductPriceSchema;