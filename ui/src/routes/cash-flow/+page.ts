import { graphql, query } from '#lib/graphql/client.js';
import type { PageLoad } from './$types';

const CashFlowQuery = graphql(`
	query CashFlow($from: Time!) {
		cashFlow(from: $from, months: 12) {
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
				unscheduledIn { amount currency }
				unscheduledOut { amount currency }
				unscheduledSaved { amount currency }
				unscheduled {
					id
					type
					time
					units
					cashDelta { amount currency }
					note
					asset { id displayName kind }
					counterpart { id cashAccount { id displayName } }
					schedule { id displayName }
				}
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

/** The year shown, from ?year=; defaults to the current one. Next year is allowed: it is all expectation. */
export const load: PageLoad = async ({ fetch, url }) => {
	const current = new Date().getFullYear();
	const param = Number(url.searchParams.get('year'));
	const year = Number.isInteger(param) && param > 1900 && param <= current + 1 ? param : current;
	const { cashFlow } = await query(CashFlowQuery, { from: new Date(Date.UTC(year, 0, 1)) }, fetch);
	return { year, cashFlow };
};
