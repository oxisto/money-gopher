/* eslint-disable */
import { DocumentTypeDecoration } from '@graphql-typed-document-node/core';
export type Maybe<T> = T | null;
export type InputMaybe<T> = Maybe<T>;
export type Exact<T extends { [key: string]: unknown }> = { [K in keyof T]: T[K] };
export type MakeOptional<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]?: Maybe<T[SubKey]> };
export type MakeMaybe<T, K extends keyof T> = Omit<T, K> & { [SubKey in K]: Maybe<T[SubKey]> };
export type MakeEmpty<T extends { [key: string]: unknown }, K extends keyof T> = { [_ in K]?: never };
export type Incremental<T> = T | { [P in keyof T]?: P extends ' $fragmentName' | '__typename' ? T[P] : never };
/** All built-in and custom scalars, mapped to their actual values */
export type Scalars = {
  ID: { input: string; output: string; }
  String: { input: string; output: string; }
  Boolean: { input: boolean; output: boolean; }
  Int: { input: number; output: number; }
  Float: { input: number; output: number; }
  Time: { input: any; output: any; }
};

export type Account = {
  __typename?: 'Account';
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  referenceAccount?: Maybe<Account>;
  type: AccountType;
};

export type AccountInput = {
  displayName: Scalars['String']['input'];
  id: Scalars['String']['input'];
  type: AccountType;
};

export enum AccountType {
  Bank = 'BANK',
  Brokerage = 'BROKERAGE',
  Loan = 'LOAN'
}

export type Currency = {
  __typename?: 'Currency';
  amount: Scalars['Int']['output'];
  symbol: Scalars['String']['output'];
};

export type CurrencyInput = {
  amount: Scalars['Int']['input'];
  symbol: Scalars['String']['input'];
};

export type ListedSecurity = {
  __typename?: 'ListedSecurity';
  currency: Scalars['String']['output'];
  latestQuote?: Maybe<Currency>;
  latestQuoteTimestamp?: Maybe<Scalars['Time']['output']>;
  security: Security;
  ticker: Scalars['String']['output'];
};

export type ListedSecurityInput = {
  currency: Scalars['String']['input'];
  ticker: Scalars['String']['input'];
};

export type Mutation = {
  __typename?: 'Mutation';
  createAccount: Account;
  createPortfolio: Portfolio;
  createSecurity: Security;
  createTransaction: Transaction;
  deleteAccount: Account;
  /**
   * Triggers a quote update for the given security IDs. If no security IDs are
   * provided, all securities will be updated.
   */
  triggerQuoteUpdate: Array<Maybe<ListedSecurity>>;
  updatePortfolio: Portfolio;
  updateSecurity: Security;
  updateTransaction: Transaction;
};


export type MutationCreateAccountArgs = {
  input: AccountInput;
};


export type MutationCreatePortfolioArgs = {
  input: PortfolioInput;
};


export type MutationCreateSecurityArgs = {
  input: SecurityInput;
};


export type MutationCreateTransactionArgs = {
  input: TransactionInput;
};


export type MutationDeleteAccountArgs = {
  id: Scalars['String']['input'];
};


export type MutationTriggerQuoteUpdateArgs = {
  securityIDs?: InputMaybe<Array<Scalars['String']['input']>>;
};


export type MutationUpdatePortfolioArgs = {
  id: Scalars['ID']['input'];
  input: PortfolioInput;
};


export type MutationUpdateSecurityArgs = {
  id: Scalars['ID']['input'];
  input: SecurityInput;
};


export type MutationUpdateTransactionArgs = {
  id: Scalars['String']['input'];
  input: TransactionInput;
};

export type Portfolio = {
  __typename?: 'Portfolio';
  accounts: Array<Account>;
  displayName: Scalars['String']['output'];
  events: Array<PortfolioEvent>;
  id: Scalars['String']['output'];
  snapshot?: Maybe<PortfolioSnapshot>;
};


export type PortfolioSnapshotArgs = {
  when?: InputMaybe<Scalars['Time']['input']>;
};

export type PortfolioEvent = {
  __typename?: 'PortfolioEvent';
  security?: Maybe<Security>;
  time: Scalars['Time']['output'];
  type: PortfolioEventType;
};

export enum PortfolioEventType {
  Buy = 'BUY',
  DeliveryInbound = 'DELIVERY_INBOUND',
  DeliveryOutbound = 'DELIVERY_OUTBOUND',
  DepositCash = 'DEPOSIT_CASH',
  Dividend = 'DIVIDEND',
  Sell = 'SELL',
  WithdrawCash = 'WITHDRAW_CASH'
}

export type PortfolioInput = {
  accountIds: Array<Scalars['String']['input']>;
  displayName: Scalars['String']['input'];
  id: Scalars['String']['input'];
};

export type PortfolioPosition = {
  __typename?: 'PortfolioPosition';
  amount: Scalars['Float']['output'];
  /** Gains contains the relative amount of profit or loss in this position. */
  gains: Scalars['Float']['output'];
  /**
   * MarketPrice is the current market price of this position, as retrieved from
   * the securities service.
   */
  marketPrice: Currency;
  /**
   * MarketValue is the current market value of this position, as retrieved from
   * the securities service.
   */
  marketValue: Currency;
  /** ProfitOrLoss contains the absolute amount of profit or loss in this position. */
  profitOrLoss: Currency;
  /**
   * PurchasePrice was the market price of this position when it was bought (net;
   * exclusive of any fees).
   */
  purchasePrice: Currency;
  /**
   * PurchaseValue was the market value of this position when it was bought (net;
   * exclusive of any fees).
   */
  purchaseValue: Currency;
  security: Security;
  /**
   * TotalFees is the total amount of fees accumulating in this position through
   * various transactions.
   */
  totalFees: Currency;
};

export type PortfolioSnapshot = {
  __typename?: 'PortfolioSnapshot';
  cash: Currency;
  firstTransactionTime: Scalars['Time']['output'];
  positions: Array<PortfolioPosition>;
  time: Scalars['Time']['output'];
  totalGains: Scalars['Float']['output'];
  totalMarketValue: Currency;
  totalPortfolioValue?: Maybe<Currency>;
  totalProfitOrLoss: Currency;
  totalPurchaseValue: Currency;
};

export type Query = {
  __typename?: 'Query';
  account?: Maybe<Account>;
  accounts: Array<Account>;
  portfolio?: Maybe<Portfolio>;
  portfolios: Array<Portfolio>;
  securities: Array<Security>;
  security?: Maybe<Security>;
  transactions: Array<Transaction>;
};


export type QueryAccountArgs = {
  id: Scalars['String']['input'];
};


export type QueryPortfolioArgs = {
  id: Scalars['String']['input'];
};


export type QuerySecurityArgs = {
  id: Scalars['String']['input'];
};


export type QueryTransactionsArgs = {
  accountID: Scalars['String']['input'];
};

export type Security = {
  __typename?: 'Security';
  displayName: Scalars['String']['output'];
  id: Scalars['String']['output'];
  listedAs?: Maybe<Array<ListedSecurity>>;
  quoteProvider?: Maybe<Scalars['String']['output']>;
};

export type SecurityInput = {
  displayName: Scalars['String']['input'];
  id: Scalars['String']['input'];
  listedAs?: InputMaybe<Array<ListedSecurityInput>>;
};

export type Transaction = {
  __typename?: 'Transaction';
  amount: Scalars['Float']['output'];
  destinationAccount: Account;
  fees: Currency;
  id: Scalars['String']['output'];
  price: Currency;
  security: Security;
  sourceAccount: Account;
  time: Scalars['Time']['output'];
  type: PortfolioEventType;
};

export type TransactionInput = {
  amount: Scalars['Float']['input'];
  destinationAccountID: Scalars['String']['input'];
  fees: CurrencyInput;
  price: CurrencyInput;
  securityID: Scalars['String']['input'];
  sourceAccountID: Scalars['String']['input'];
  taxes: CurrencyInput;
  time: Scalars['Time']['input'];
  type: PortfolioEventType;
};

export type ListAccountsQueryVariables = Exact<{ [key: string]: never; }>;


export type ListAccountsQuery = { __typename?: 'Query', accounts: Array<{ __typename?: 'Account', id: string, displayName: string }> };

export class TypedDocumentString<TResult, TVariables>
  extends String
  implements DocumentTypeDecoration<TResult, TVariables>
{
  __apiType?: DocumentTypeDecoration<TResult, TVariables>['__apiType'];

  constructor(private value: string, public __meta__?: Record<string, any> | undefined) {
    super(value);
  }

  toString(): string & DocumentTypeDecoration<TResult, TVariables> {
    return this.value;
  }
}

export const ListAccountsDocument = new TypedDocumentString(`
    query ListAccounts {
  accounts {
    id
    displayName
  }
}
    `) as unknown as TypedDocumentString<ListAccountsQuery, ListAccountsQueryVariables>;