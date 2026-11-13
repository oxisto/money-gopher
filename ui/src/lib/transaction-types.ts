/**
 * Shared shapes for ActivityFeed/TransactionEditForm/TransactionForm, used
 * on the activity, portfolio and cash account pages — all query the same
 * Transaction fields, just scoped differently, so the components take this
 * structural type instead of one tied to a single generated query.
 */
export interface TransactionRow {
	id: string;
	type: string;
	time: Date;
	units: number;
	price: { amount: string; currency: string } | null;
	fees: { amount: number; currency: string };
	taxes: { amount: number; currency: string };
	cashDelta: { amount: number; currency: string };
	fxRate: string | null;
	asset: { id: string; displayName: string } | null;
	cashAccount: { id: string } | null;
	/** Set on both legs of a transfer between own accounts. */
	transferID?: string | null;
	/** Booked by a schedule; not yet confirmed by a statement when document is null. */
	scheduleID?: string | null;
	schedule?: { id: string; displayName: string } | null;
	/** Free text: the statement's wording or the person's own; the title when set. */
	note?: string | null;
	counterpart?: { id: string; cashAccount: { id: string; displayName: string } | null } | null;
}

export interface AssetOption {
	id: string;
	displayName: string;
}

export interface CashAccountOption {
	id: string;
	displayName: string;
	/** CASH when missing; INTEREST on a LOAN is a charge, not a credit. */
	kind?: string | null;
}

/**
 * A TransactionRow as the activity feed shows it: with the accounts it
 * touches by name, the realized gain of a sale, and its source document.
 */
export interface FeedTransaction extends Omit<TransactionRow, 'cashAccount'> {
	gain: { amount: number; currency: string } | null;
	portfolio: { id: string; displayName: string } | null;
	cashAccount: { id: string; displayName: string } | null;
	document: { id: string; filename: string } | null;
	/** Every document describing it: the origin, and each statement that confirmed it. */
	documents?: { id: string; filename: string }[];
}

/** Every TransactionType in api/schema.graphql, in its order. */
export const TRANSACTION_TYPES = [
	'BUY',
	'SELL',
	'DELIVERY_INBOUND',
	'DELIVERY_OUTBOUND',
	'DIVIDEND',
	'INTEREST',
	'DEPOSIT_CASH',
	'WITHDRAW_CASH',
	'ACCOUNT_FEES',
	'TAX_REFUND'
] as const;
export type TransactionType = (typeof TRANSACTION_TYPES)[number];
