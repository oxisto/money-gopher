export function currency(value: number, currency: string): string {
	const formatter = Intl.NumberFormat('de', {
		style: 'currency',
		currency: currency
	});

	return formatter.format(value);
}
