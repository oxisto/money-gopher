/** What an asset is, as the API's AssetKind enum, with its label. */
export const assetKinds = [
	['STOCK', 'Stock'],
	['ETF', 'ETF'],
	['FUND', 'Fund'],
	['BOND', 'Bond'],
	['PROPERTY', 'Property'],
	['VEHICLE', 'Vehicle'],
	['LOAN', 'Loan'],
	['OTHER', 'Other']
] as const;

export type AssetKind = (typeof assetKinds)[number][0];

export const kindLabel = Object.fromEntries(assetKinds) as Record<AssetKind, string>;

/** Narrows a select's string value to the enum; undefined when unset. */
export function asKind(value: string | null | undefined): AssetKind | undefined {
	return assetKinds.some(([kind]) => kind === value) ? (value as AssetKind) : undefined;
}

/**
 * Kinds that no quote feed knows: their price series is valued by hand.
 * The office room, the car, a loan — and "other", which is what it says.
 */
export function isValuedByHand(kind: string | null | undefined): boolean {
	return kind === 'PROPERTY' || kind === 'VEHICLE' || kind === 'LOAN' || kind === 'OTHER';
}

/** The kinds a personal asset can be: those valued by hand. */
export const handValuedKinds = assetKinds.filter(([kind]) => isValuedByHand(kind));

/** A ticker-shaped name for a manual price series: "Office room" → "OFFICE-ROOM". */
export function tickerFor(name: string): string {
	return (
		name
			.normalize('NFD')
			.replace(/[̀-ͯ]/g, '')
			.toUpperCase()
			.replace(/[^A-Z0-9]+/g, '-')
			.replace(/^-|-$/g, '')
			.slice(0, 24) || 'ASSET'
	);
}
