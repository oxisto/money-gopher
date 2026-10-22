import { graphql, query } from '#lib/graphql/client.js';
import type { PageLoad } from './$types';

const TransactionDetail = graphql(`
	query TransactionDetail($id: ID!) {
		transaction(id: $id) {
			id
			type
			time
			units
			price { amount currency }
			fees { amount currency }
			taxes { amount currency }
			cashDelta { amount currency }
			gain { amount currency }
			fxRate
			source
			note
			portfolio { id displayName }
			cashAccount { id displayName }
			asset {
				id
				displayName
				identifiers { kind value }
			}
			documents {
				id
				filename
				contentType
				state
				detectedBank
				createdAt
			}
		}
		securities { id displayName }
		assets { id displayName }
		cashAccounts { id displayName }
	}
`);

export const load: PageLoad = async ({ fetch, params }) =>
	query(TransactionDetail, { id: params.id }, fetch);
