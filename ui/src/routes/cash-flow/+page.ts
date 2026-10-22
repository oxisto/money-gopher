import { graphql, query } from '#lib/graphql/client.js';
import type { PageLoad } from './$types';

const CashFlowQuery = graphql(`
	query CashFlow {
		cashFlow {
			currency
			from
			months {
				month
				expectedIn { amount currency }
				expectedOut { amount currency }
				expectedSaved { amount currency }
				expectedFree { amount currency }
				actualIn { amount currency }
				actualOut { amount currency }
				actualSaved { amount currency }
				actualFree { amount currency }
				flows {
					date
					amount { amount currency }
					booked
					schedule { id displayName type }
				}
			}
			schedules {
				id
				displayName
				type
				interval
				dayOfMonth
				nextDate
				books
				currentAmount { amount currency }
				cashAccount { id displayName kind }
				counterpartAccount { id displayName kind }
			}
		}
	}
`);

export const load: PageLoad = async ({ fetch }) => query(CashFlowQuery, {}, fetch);
