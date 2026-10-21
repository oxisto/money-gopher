import { graphql, query } from '#lib/graphql/client.js';
import { ActivityTransaction } from '#lib/graphql/documents.js';
import type { PageLoad } from './$types';

const Activity = graphql(
	`
		query Activity {
			portfolios {
				id
				transactions { ...ActivityTransaction }
			}
			cashAccounts {
				id
				displayName
				currency
				transactions { ...ActivityTransaction }
			}
			securities { id displayName }
		}
	`,
	[ActivityTransaction]
);

export const load: PageLoad = async ({ fetch }) => query(Activity, {}, fetch);
