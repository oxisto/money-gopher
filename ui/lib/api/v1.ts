/**
 * This file was auto-generated by openapi-typescript.
 * Do not make direct changes to the file.
 */

export interface paths {
    "/v1/portfolios": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["PortfolioService_ListPortfolios"];
        put?: never;
        post: operations["PortfolioService_CreatePortfolio"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/portfolios/{name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["PortfolioService_GetPortfolio"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/portfolios/{portfolioName}/snapshot": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["PortfolioService_GetPortfolioSnapshot"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/portfolios/{portfolioName}/transactions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["PortfolioService_ListPortfolioTransactions"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/portfolios/{transaction.portfolio_name}/transactions": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put?: never;
        post: operations["PortfolioService_CreatePortfolioTransaction"];
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/securities": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["SecuritiesService_ListSecurities"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/securities/{name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["SecuritiesService_GetSecurity"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/transactions/{name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get: operations["PortfolioService_GetPortfolioTransaction"];
        put?: never;
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
    "/v1/transactions/{transaction.name}": {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        get?: never;
        put: operations["PortfolioService_UpdatePortfolioTransaction"];
        post?: never;
        delete?: never;
        options?: never;
        head?: never;
        patch?: never;
        trace?: never;
    };
}
export type webhooks = Record<string, never>;
export interface components {
    schemas: {
        /** @description Currency is a currency value in the lowest unit of the selected currency
         *      (e.g., cents for EUR/USD). */
        Currency: {
            /** Format: int32 */
            value: number;
            symbol: string;
        };
        /** @description Contains an arbitrary serialized message along with a @type that describes the type of the serialized message. */
        GoogleProtobufAny: {
            /** @description The type of the serialized message. */
            "@type"?: string;
        } & {
            [key: string]: unknown;
        };
        ListPortfolioTransactionsResponse: {
            transactions: components["schemas"]["PortfolioEvent"][];
        };
        ListPortfoliosResponse: {
            portfolios: components["schemas"]["Portfolio"][];
        };
        ListSecuritiesResponse: {
            securities: components["schemas"]["Security"][];
        };
        ListedSecurity: {
            securityId: string;
            ticker: string;
            currency: string;
            latestQuote?: components["schemas"]["Currency"];
            /** Format: date-time */
            latestQuoteTimestamp?: string;
        };
        Portfolio: {
            name: string;
            displayName: string;
            /** @description BankAccountName contains the name/identifier of the underlying bank
             *      account. */
            bankAccountName: string;
            /** @description Events contains all portfolio events, such as buy/sell transactions,
             *      dividends or other. They need to be ordered by time (ascending). */
            events?: components["schemas"]["PortfolioEvent"][];
        };
        PortfolioEvent: {
            name: string;
            /**
             * Format: enum
             * @enum {string}
             */
            type: "PORTFOLIO_EVENT_TYPE_UNSPECIFIED" | "PORTFOLIO_EVENT_TYPE_BUY" | "PORTFOLIO_EVENT_TYPE_SELL" | "PORTFOLIO_EVENT_TYPE_DELIVERY_INBOUND" | "PORTFOLIO_EVENT_TYPE_DELIVERY_OUTBOUND" | "PORTFOLIO_EVENT_TYPE_DIVIDEND" | "PORTFOLIO_EVENT_TYPE_INTEREST" | "PORTFOLIO_EVENT_TYPE_DEPOSIT_CASH" | "PORTFOLIO_EVENT_TYPE_WITHDRAW_CASH" | "PORTFOLIO_EVENT_TYPE_ACCOUNT_FEES" | "PORTFOLIO_EVENT_TYPE_TAX_REFUND";
            /** Format: date-time */
            time: string;
            portfolioName: string;
            securityId: string;
            /** Format: double */
            amount: number;
            price: components["schemas"]["Currency"];
            fees: components["schemas"]["Currency"];
            taxes: components["schemas"]["Currency"];
        };
        PortfolioPosition: {
            security: components["schemas"]["Security"];
            /** Format: double */
            amount: number;
            /** @description PurchaseValue was the market value of this position when it was bought
             *      (net; exclusive of any fees). */
            purchaseValue: components["schemas"]["Currency"];
            /** @description PurchasePrice was the market price of this position when it was bought
             *      (net; exclusive of any fees). */
            purchasePrice: components["schemas"]["Currency"];
            /** @description MarketValue is the current market value of this position, as retrieved from
             *      the securities service. */
            marketValue: components["schemas"]["Currency"];
            /** @description MarketPrice is the current market price of this position, as retrieved from
             *      the securities service. */
            marketPrice: components["schemas"]["Currency"];
            /** @description TotalFees is the total amount of fees accumulating in this position through
             *      various transactions. */
            totalFees: components["schemas"]["Currency"];
            /** @description ProfitOrLoss contains the absolute amount of profit or loss in this
             *      position. */
            profitOrLoss: components["schemas"]["Currency"];
            /**
             * Format: double
             * @description Gains contains the relative amount of profit or loss in this position.
             */
            gains: number;
        };
        /** @description PortfolioSnapshot represents a snapshot in time of the portfolio. It can for
         *      example be the current state of the portfolio but also represent the state of
         *      the portfolio at a certain time in the past. */
        PortfolioSnapshot: {
            /**
             * Format: date-time
             * @description Time is the time when this snapshot was taken.
             */
            time: string;
            /** @description Positions holds the current positions within the snapshot and their value. */
            positions: {
                [key: string]: components["schemas"]["PortfolioPosition"];
            };
            /**
             * Format: date-time
             * @description FirstTransactionTime is the time of the first transaction with the
             *      snapshot.
             */
            firstTransactionTime: string;
            /** @description TotalPurchaseValue contains the total purchase value of all asset positions */
            totalPurchaseValue: components["schemas"]["Currency"];
            /** @description TotalMarketValue contains the total market value of all asset positions */
            totalMarketValue: components["schemas"]["Currency"];
            /** @description TotalProfitOrLoss contains the total absolute amount of profit or loss in
             *      this snapshot, based on asset value. */
            totalProfitOrLoss: components["schemas"]["Currency"];
            /**
             * Format: double
             * @description TotalGains contains the total relative amount of profit or loss in this
             *      snapshot, based on asset value.
             */
            totalGains: number;
            /** @description Cash contains the current amount of cash in the portfolio's bank
             *      account(s). */
            cash: components["schemas"]["Currency"];
            /** @description TotalPortfolioValue contains the amount of cash plus the total market value
             *      of all assets. */
            totalPortfolioValue: components["schemas"]["Currency"];
        };
        Security: {
            /** @description Name contains the unique resource name. For a stock or bond, this should be
             *      an ISIN. */
            name: string;
            /** @description DisplayName contains the human readable name. */
            displayName: string;
            listedOn: components["schemas"]["ListedSecurity"][];
            quoteProvider: string;
        };
        /** @description The `Status` type defines a logical error model that is suitable for different programming environments, including REST APIs and RPC APIs. It is used by [gRPC](https://github.com/grpc). Each `Status` message contains three pieces of data: error code, error message, and error details. You can find out more about this error model and how to work with it in the [API Design Guide](https://cloud.google.com/apis/design/errors). */
        Status: {
            /**
             * Format: int32
             * @description The status code, which should be an enum value of [google.rpc.Code][google.rpc.Code].
             */
            code?: number;
            /** @description A developer-facing error message, which should be in English. Any user-facing error message should be localized and sent in the [google.rpc.Status.details][google.rpc.Status.details] field, or localized by the client. */
            message?: string;
            /** @description A list of messages that carry the error details.  There is a common set of message types for APIs to use. */
            details?: components["schemas"]["GoogleProtobufAny"][];
        };
    };
    responses: never;
    parameters: never;
    requestBodies: never;
    headers: never;
    pathItems: never;
}
export type SchemaCurrency = components['schemas']['Currency'];
export type SchemaGoogleProtobufAny = components['schemas']['GoogleProtobufAny'];
export type SchemaListPortfolioTransactionsResponse = components['schemas']['ListPortfolioTransactionsResponse'];
export type SchemaListPortfoliosResponse = components['schemas']['ListPortfoliosResponse'];
export type SchemaListSecuritiesResponse = components['schemas']['ListSecuritiesResponse'];
export type SchemaListedSecurity = components['schemas']['ListedSecurity'];
export type SchemaPortfolio = components['schemas']['Portfolio'];
export type SchemaPortfolioEvent = components['schemas']['PortfolioEvent'];
export type SchemaPortfolioPosition = components['schemas']['PortfolioPosition'];
export type SchemaPortfolioSnapshot = components['schemas']['PortfolioSnapshot'];
export type SchemaSecurity = components['schemas']['Security'];
export type SchemaStatus = components['schemas']['Status'];
export type $defs = Record<string, never>;
export interface operations {
    PortfolioService_ListPortfolios: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ListPortfoliosResponse"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_CreatePortfolio: {
        parameters: {
            query?: never;
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["Portfolio"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Portfolio"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_GetPortfolio: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                name: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Portfolio"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_GetPortfolioSnapshot: {
        parameters: {
            query?: {
                /** @description Time is the point in time of the requested snapshot. */
                time?: string;
            };
            header?: never;
            path: {
                /** @description PortfolioName is the name / identifier of the portfolio we want to
                 *      "snapshot". */
                portfolioName: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PortfolioSnapshot"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_ListPortfolioTransactions: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                portfolioName: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ListPortfolioTransactionsResponse"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_CreatePortfolioTransaction: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                "transaction.portfolio_name": string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["PortfolioEvent"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PortfolioEvent"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    SecuritiesService_ListSecurities: {
        parameters: {
            query?: {
                "filter.securityIds"?: string[];
            };
            header?: never;
            path?: never;
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["ListSecuritiesResponse"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    SecuritiesService_GetSecurity: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                name: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Security"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_GetPortfolioTransaction: {
        parameters: {
            query?: never;
            header?: never;
            path: {
                name: string;
            };
            cookie?: never;
        };
        requestBody?: never;
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PortfolioEvent"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
    PortfolioService_UpdatePortfolioTransaction: {
        parameters: {
            query?: {
                updateMask?: string;
            };
            header?: never;
            path: {
                "transaction.name": string;
            };
            cookie?: never;
        };
        requestBody: {
            content: {
                "application/json": components["schemas"]["PortfolioEvent"];
            };
        };
        responses: {
            /** @description OK */
            200: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["PortfolioEvent"];
                };
            };
            /** @description Default error response */
            default: {
                headers: {
                    [name: string]: unknown;
                };
                content: {
                    "application/json": components["schemas"]["Status"];
                };
            };
        };
    };
}
