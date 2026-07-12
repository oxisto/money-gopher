/**
 * A monetary value as it comes out of the GraphQL API: minor units (e.g.
 * euro cents) plus an ISO 4217 currency code.
 */
export interface Money {
	amount: number;
	currency: string;
}

/** Formats a Money value using the browser locale, e.g. "1.234,56 €". */
export function formatMoney(money: Money): string {
	return new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: money.currency
	}).format(money.amount / 100);
}

/**
 * Parses a user-entered decimal amount (e.g. "1234.56") into minor units.
 * Returns null for empty or invalid input.
 */
export function parseAmount(input: string): number | null {
	const value = parseFloat(input.replace(',', '.'));
	if (isNaN(value)) {
		return null;
	}
	return Math.round(value * 100);
}

/** Formats a transaction timestamp using the browser locale. */
export function formatDate(date: Date): string {
	return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(date);
}

/** Formats a fraction as a signed percentage, e.g. 0.05 → "+5.0%". */
export function formatPercent(fraction: number): string {
	return new Intl.NumberFormat(undefined, {
		style: 'percent',
		minimumFractionDigits: 1,
		maximumFractionDigits: 1,
		signDisplay: 'exceptZero'
	}).format(fraction);
}
