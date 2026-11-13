import { graphql, query } from '#lib/graphql/client.js';
import { ActivityTransaction, HoldingPosition } from '#lib/graphql/documents.js';
import type { PageLoad } from './$types';

const ReturnsFields = graphql(`
	fragment ReturnsFields on Returns @_unmask {
		from
		to
		realized { amount currency }
		income { amount currency }
		costs { amount currency }
		valueChange { amount currency }
		total { amount currency }
		timeWeightedReturn
	}
`);

const PortfolioDetail = graphql(
	`
		query PortfolioDetail($id: ID!) {
			portfolio(id: $id) {
				id
				displayName
				cashAccount { id displayName }
				snapshot {
					totalPurchaseValue { amount currency }
					totalMarketValue { amount currency }
					totalProfitOrLoss { amount currency }
					totalGains
					performance(period: ALL_TIME) { timeWeightedReturn }
					ytd: returns(period: YEAR_TO_DATE) { ...ReturnsFields }
					oneYear: returns(period: ONE_YEAR) { ...ReturnsFields }
					allTime: returns(period: ALL_TIME) { ...ReturnsFields }
					returnsByYear { ...ReturnsFields }
					positions { ...HoldingPosition }
				}
				transactions { ...ActivityTransaction }
			}
			securities { id displayName }
			assets { id displayName }
			cashAccounts { id displayName currency kind }
		}
	`,
	[HoldingPosition, ActivityTransaction, ReturnsFields]
);

export const load: PageLoad = async ({ fetch, params }) =>
	query(PortfolioDetail, { id: params.id }, fetch);
