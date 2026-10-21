import { graphql, query } from '#lib/graphql/client.js';
import { HoldingPosition } from '#lib/graphql/documents.js';
import type { PageLoad } from './$types';

const Holdings = graphql(
	`
		query Holdings {
			portfolios {
				id
				displayName
				snapshot {
					totalMarketValue { amount currency }
					totalPurchaseValue { amount currency }
					totalProfitOrLoss { amount currency }
					performance(period: ALL_TIME) { timeWeightedReturn }
					positions { ...HoldingPosition }
				}
			}
		}
	`,
	[HoldingPosition]
);

export const load: PageLoad = async ({ fetch }) => query(Holdings, {}, fetch);
