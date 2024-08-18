import { Currency } from "@/lib/gen/mgo_pb";

export function classNames(...classes: string[]) {
  return classes.filter(Boolean).join(" ");
}

export function currency(c: Currency | undefined): string {
	if (c === undefined) {
		return '';
	}

	const formatter = Intl.NumberFormat(navigator.language, {
		style: 'currency',
		currency: c.symbol
	});

	return formatter.format(c.value / 100);
}

export function shorten(text: string): string {
	let max = 30;

	if (text.length > max) {
		return text.substring(0, max) + '...';
	} else {
		return text;
	}
}