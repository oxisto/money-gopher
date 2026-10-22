import { graphql, type ResultOf } from './client';

// Documents used by more than one route or by components that need their
// result types. A page's own query lives in its +page.ts instead.

// @_unmask: the fields are read directly where the fragment is spread,
// rather than through readFragment().

/** Everything ActivityFeed shows and TransactionEditForm edits. */
export const ActivityTransaction = graphql(`
	fragment ActivityTransaction on Transaction @_unmask {
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
		asset { id displayName }
		portfolio { id displayName }
		cashAccount { id displayName }
		document { id filename }
		transferID
		counterpart { id cashAccount { id displayName } }
		scheduleID
	}
`);

/** A recurring cash flow as the account page's panel and form need it. */
export const ScheduleFields = graphql(`
	fragment ScheduleFields on Schedule @_unmask {
		id
		displayName
		type
		interval
		dayOfMonth
		start
		end
		books
		booked
		cashAccount { id displayName }
		counterpartAccount { id displayName }
		asset { id displayName }
		steps { from amount { amount currency } }
		currentAmount { amount currency }
		nextDate
		unconfirmed { id }
	}
`);

/** A portfolio position as HoldingsTable and the holding page need it. */
export const HoldingPosition = graphql(`
	fragment HoldingPosition on Position @_unmask {
		asset {
			id
			displayName
			kind
			identifiers { kind value }
		}
		units
		purchasePrice { amount currency }
		marketPrice { amount currency }
		purchaseValue { amount currency }
		marketValue { amount currency }
		profitOrLoss { amount currency }
	}
`);

/** The shell: person switcher, account list and the import badge. */
export const LayoutQuery = graphql(`
	query Layout {
		me {
			id
			displayName
			activePerson { id displayName }
			persons { id displayName }
		}
		portfolios {
			id
			displayName
			snapshot { totalMarketValue { amount currency } }
		}
		cashAccounts {
			id
			displayName
			kind estimated balance { amount currency }
		}
		documents(state: PARSED) { id }
	}
`);
export type LayoutResult = ResultOf<typeof LayoutQuery>;

/** The person's own assets with their position, for the Assets page. */
export const AssetsQuery = graphql(
	`
		query Assets {
			assets {
				id
				displayName
				kind
				listings {
					id
					currency
					latestQuote {
						time
						price { amount currency }
					}
				}
			}
			portfolios {
				id
				displayName
				snapshot {
					positions { ...HoldingPosition }
				}
			}
			cashAccounts { id displayName currency }
		}
	`,
	[HoldingPosition]
);
export type AssetsResult = ResultOf<typeof AssetsQuery>;

export const SecuritiesQuery = graphql(`
	query Securities {
		securities {
			id
			displayName
			kind
			identifiers { kind value }
			underlyingAsset { id displayName }
			listings {
				id
				ticker
				exchange
				currency
				quoteProvider
				latestQuote {
					time
					price { amount currency }
				}
			}
		}
	}
`);
export type SecuritiesResult = ResultOf<typeof SecuritiesQuery>;

export const PeopleQuery = graphql(`
	query People {
		persons {
			id
			displayName
			users { id displayName }
		}
		users { id displayName }
	}
`);
export type PeopleResult = ResultOf<typeof PeopleQuery>;
