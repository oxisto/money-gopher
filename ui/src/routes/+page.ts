import { graphql, query, type VariablesOf } from '#lib/graphql/client.js';
import type { PageLoad } from './$types';

const OverviewTransaction = graphql(`
	fragment OverviewTransaction on Transaction @_unmask {
		id
		type
		time
		units
		cashDelta { amount currency }
		gain { amount currency }
		asset { displayName }
		portfolio { displayName }
		cashAccount { displayName }
		transferID
		counterpart { id cashAccount { id displayName } }
		schedule { id displayName }
	}
`);

const Dashboard = graphql(`
	query Dashboard {
		portfolios {
			id
			displayName
			snapshot {
				totalMarketValue { amount currency }
				totalPurchaseValue { amount currency }
				totalProfitOrLoss { amount currency }
				totalGains
				performance(period: YEAR_TO_DATE) { timeWeightedReturn }
				positions {
					asset { id displayName }
					marketValue { amount currency }
				}
			}
			transactions { ...OverviewTransaction }
		}
		cashAccounts {
			id
			displayName
			kind
			estimated
			currency
			balance { amount currency }
			transactions { ...OverviewTransaction }
		}
		realizedGains {
			dividends { amount currency }
			interest { amount currency }
			tradingGains { amount currency }
			total { amount currency }
			entries {
				id
				time
				type
				amount { amount currency }
			}
		}
		documents(state: PARSED) {
			id
			filename
			transactionDate
			suggestedCashAccount { displayName }
			stagedTransactions {
				type
				units
				cashDelta { amount currency }
				calculatedCashDelta { amount currency }
				assetHint
				asset { displayName }
			}
		}
	}

`, [OverviewTransaction]);

const NetWorthHistory = graphql(`
	query NetWorthHistory($period: Period!) {
		netWorthHistory(period: $period, points: 90) {
			time
			netWorth { amount currency }
			invested { amount currency }
			estimated { amount currency }
		}
	}
`);

/** The chart's ranges, as they appear in the URL (?range=1Y). */
const ranges: Record<string, VariablesOf<typeof NetWorthHistory>['period']> = {
	'1M': 'ONE_MONTH',
	'6M': 'SIX_MONTHS',
	YTD: 'YEAR_TO_DATE',
	'1Y': 'ONE_YEAR',
	Max: 'ALL_TIME'
};

export const load: PageLoad = async ({ fetch, url }) => {
	const param = url.searchParams.get('range') ?? 'YTD';
	const range = param in ranges ? param : 'YTD';
	const [dashboard, history] = await Promise.all([
		query(Dashboard, {}, fetch),
		query(NetWorthHistory, { period: ranges[range] }, fetch)
	]);
	return { dashboard, history: history.netWorthHistory, range };
};
