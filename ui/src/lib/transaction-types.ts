/**
 * Shared shapes for TransactionTable/TransactionEditForm/TransactionForm,
 * used both on a portfolio's detail page and a cash account's — both pages
 * query the same Transaction fields, just scoped differently, so the
 * components take this structural type instead of one tied to a single
 * generated query.
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
	security: { id: string; displayName: string } | null;
	cashAccount: { id: string } | null;
	document: { id: string; filename: string } | null;
}

export interface SecurityOption {
	id: string;
	displayName: string;
}

export interface CashAccountOption {
	id: string;
	displayName: string;
}
