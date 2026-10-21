import { graphql, query } from '$lib/graphql/client';
import { ActivityTransaction, HoldingPosition } from '$lib/graphql/documents';
import type { PageLoad } from './$types';

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
					positions { ...HoldingPosition }
				}
				transactions { ...ActivityTransaction }
			}
			securities { id displayName }
			cashAccounts { id displayName currency }
		}
	`,
	[HoldingPosition, ActivityTransaction]
);

export const load: PageLoad = async ({ fetch, params }) =>
	query(PortfolioDetail, { id: params.id }, fetch);
