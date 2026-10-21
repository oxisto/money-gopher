import { graphql, query } from '$lib/graphql/client';
import type { PageLoad } from './$types';

const RealizedGains = graphql(`
	query RealizedGains($from: Time, $to: Time) {
		realizedGains(from: $from, to: $to) {
			from
			to
			currency
			dividends { amount currency }
			interest { amount currency }
			tradingGains { amount currency }
			total { amount currency }
			entries {
				id
				time
				type
				amount { amount currency }
				security { id displayName }
				cashAccount { displayName }
				portfolio { displayName }
			}
		}
	}
`);

/** The year shown, from ?year=; defaults to the current one. */
export const load: PageLoad = async ({ fetch, url }) => {
	const current = new Date().getFullYear();
	const param = Number(url.searchParams.get('year'));
	const year = Number.isInteger(param) && param > 1900 && param <= current ? param : current;
	// Exclusive upper bound: the instant January 1 of the following year
	// begins, so December 31 is fully included.
	const from = new Date(Date.UTC(year, 0, 1));
	const to = new Date(Date.UTC(year + 1, 0, 1));
	const { realizedGains } = await query(RealizedGains, { from, to }, fetch);
	return { year, gains: realizedGains };
};
