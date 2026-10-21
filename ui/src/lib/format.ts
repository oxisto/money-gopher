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

/**
 * A per-unit price, exact to more than 2 decimals (e.g. "8.2656") — unlike
 * Money, amount is a Decimal string, never a number, so it never passes
 * through float64 on its way to or from the API.
 */
export interface Price {
	amount: string;
	currency: string;
}

/**
 * Formats a Price value using the browser locale, e.g. "8,2656 €". Exact
 * to 8 decimals by default; pass fewer for derived prices like an average
 * cost, whose long tails ("99.99972727") are noise, not information.
 */
export function formatPrice(price: Price, maxDecimals = 8): string {
	return new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: price.currency,
		minimumFractionDigits: 2,
		maximumFractionDigits: maxDecimals
	}).format(Number(price.amount));
}

/**
 * Parses a user-entered decimal price (e.g. "8,2656") into the canonical
 * dot-decimal string the Decimal scalar expects. Unlike parseAmount, this
 * never multiplies or rounds — the exact digits the user typed are what gets
 * sent. Returns null for empty or invalid input.
 */
export function parsePrice(input: string): string | null {
	const trimmed = input.trim().replace(',', '.');
	if (trimmed === '' || isNaN(Number(trimmed))) {
		return null;
	}
	return trimmed;
}

/**
 * Formats a Money value with an explicit sign, e.g. "+€12.50" or "−€3.00",
 * using a real minus sign so columns of gains and losses line up.
 */
export function formatSignedMoney(money: Money): string {
	const abs = formatMoney({ ...money, amount: Math.abs(money.amount) });
	return money.amount > 0 ? `+${abs}` : money.amount < 0 ? `−${abs}` : abs;
}

/** The text color token for a gain (+), a loss (−) or neither. */
export function toneOf(amount: number): string {
	return amount > 0 ? 'text-pos' : amount < 0 ? 'text-neg' : 'text-muted';
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

/** Formats a Money value compactly for tight spaces, e.g. "€342.1K". */
export function formatMoneyCompact(money: Money): string {
	return new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: money.currency,
		notation: 'compact',
		maximumFractionDigits: 1
	}).format(money.amount / 100);
}
