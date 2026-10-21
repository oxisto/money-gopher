import { graphql, query } from '$lib/graphql/client';
import { ActivityTransaction } from '$lib/graphql/documents';
import type { PageLoad } from './$types';

const CashAccountDetail = graphql(
	`
		query CashAccountDetail($id: ID!) {
			cashAccount(id: $id) {
				id
				displayName
				currency
				iban
				balance { amount currency }
				transactions { ...ActivityTransaction }
			}
			securities { id displayName }
			cashAccounts { id displayName currency }
		}
	`,
	[ActivityTransaction]
);

export const load: PageLoad = async ({ fetch, params }) =>
	query(CashAccountDetail, { id: params.id }, fetch);
