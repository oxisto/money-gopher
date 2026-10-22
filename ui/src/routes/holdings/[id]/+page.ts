import { graphql, query } from '#lib/graphql/client.js';
import { HoldingPosition } from '#lib/graphql/documents.js';
import type { PageLoad } from './$types';

const HoldingDetail = graphql(
	`
		query HoldingDetail($id: ID!) {
			security(id: $id) {
				id
				displayName
				kind
				owner { id }
				identifiers { kind value }
				underlyingSecurity { id displayName }
				listings {
					id
					exchange
					ticker
					currency
					quoteProvider
					latestQuote {
						time
						price { amount currency }
					}
					quotes {
						time
						price { amount currency }
					}
				}
			}
			portfolios {
				id
				displayName
				snapshot {
					totalMarketValue { amount currency }
					positions { ...HoldingPosition }
				}
				transactions {
					id
					type
					time
					units
					price { amount currency }
					fees { amount currency }
					taxes { amount currency }
					cashDelta { amount currency }
					gain { amount currency }
					security { id displayName }
				}
			}
		}
	`,
	[HoldingPosition]
);

export const load: PageLoad = async ({ fetch, params }) =>
	query(HoldingDetail, { id: params.id }, fetch);
