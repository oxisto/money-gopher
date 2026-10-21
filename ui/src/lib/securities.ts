/** What a security is, as the API's SecurityKind enum, with its label. */
export const securityKinds = [
	['STOCK', 'Stock'],
	['ETF', 'ETF'],
	['FUND', 'Fund'],
	['BOND', 'Bond'],
	['OTHER', 'Other']
] as const;

export type SecurityKind = (typeof securityKinds)[number][0];

export const kindLabel = Object.fromEntries(securityKinds) as Record<SecurityKind, string>;

/** Narrows a select's string value to the enum; undefined when unset. */
export function asKind(value: string | null | undefined): SecurityKind | undefined {
	return securityKinds.some(([kind]) => kind === value) ? (value as SecurityKind) : undefined;
}
