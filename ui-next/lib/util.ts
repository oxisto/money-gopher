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
