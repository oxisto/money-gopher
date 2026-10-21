import { graphql, query } from '#lib/graphql/client.js';
import type { PageLoad } from './$types';

const Imports = graphql(`
	query Imports {
		documents {
			id
			filename
			state
			detectedBank
			error
			transactionDate
			suggestedCashAccount { id displayName currency }
			transactions {
				id
				type
				cashDelta { amount currency }
			}
			stagedTransactions {
				id
				type
				time
				units
				price { amount currency }
				fees { amount currency }
				taxes { amount currency }
				cashDelta { amount currency }
				calculatedCashDelta { amount currency }
				fxRate
				currency
				securityHint
				isin
				securityKindHint
				security { id displayName }
			}
		}
		portfolios { id displayName }
		cashAccounts { id displayName currency iban }
		securities { id displayName }
	}
`);

export const load: PageLoad = async ({ fetch }) => query(Imports, {}, fetch);
