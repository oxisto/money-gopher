import type { IconName } from './icons';

interface DescribableTransaction {
	type: string;
	units: number;
	security: { displayName: string } | null;
}

/**
 * A plain-language title and an icon for a transaction, e.g. "Bought 25 ×
 * Allianz SE" — what activity lists show instead of the raw type code.
 */
export function describeTransaction(tx: DescribableTransaction): { title: string; icon: IconName } {
	const name = tx.security?.displayName ?? 'unknown security';
	const units = formatUnits(tx.units);
	switch (tx.type) {
		case 'BUY':
			return { title: `Bought ${units} × ${name}`, icon: 'plus' };
		case 'SELL':
			return { title: `Sold ${units} × ${name}`, icon: 'minus' };
		case 'DELIVERY_INBOUND':
			return { title: `Received ${units} × ${name}`, icon: 'transferIn' };
		case 'DELIVERY_OUTBOUND':
			return { title: `Transferred out ${units} × ${name}`, icon: 'transferOut' };
		case 'DIVIDEND':
			return { title: `Dividend · ${name}`, icon: 'coin' };
		case 'INTEREST':
			return { title: 'Interest', icon: 'percent' };
		case 'DEPOSIT_CASH':
			return { title: 'Deposit', icon: 'arrowIn' };
		case 'WITHDRAW_CASH':
			return { title: 'Withdrawal', icon: 'arrowOut' };
		case 'ACCOUNT_FEES':
			return { title: 'Account fees', icon: 'receipt' };
		case 'TAX_REFUND':
			return { title: 'Tax refund', icon: 'refund' };
		default:
			return { title: tx.type, icon: 'document' };
	}
}

/** Units without trailing zeros: 25, 4.71, 0.0532. */
export function formatUnits(units: number): string {
	return new Intl.NumberFormat(undefined, { maximumFractionDigits: 6 }).format(units);
}
