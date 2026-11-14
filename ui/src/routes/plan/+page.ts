import { graphql, query } from "#lib/graphql/client.js";
import type { PageLoad } from "./$types";

const PlansQuery = graphql(`
  query Plans($from: Time!) {
    plans {
      id
      displayName
      start
      horizonMonths
      annualCost
      startingAmount {
        amount
        currency
      }
      startingAmountDeclared
      contribution {
        amount
        currency
      }
      contributionDeclared
      assumptions {
        bucket
        value {
          amount
          currency
        }
        weight
        weightDeclared
        annualReturn
        volatility
      }
      projection {
        month
        expected {
          amount
          currency
        }
        contributed {
          amount
          currency
        }
        p10 {
          amount
          currency
        }
        p25 {
          amount
          currency
        }
        p50 {
          amount
          currency
        }
        p75 {
          amount
          currency
        }
        p90 {
          amount
          currency
        }
      }
      actual {
        time
        netWorth {
          amount
          currency
        }
      }
      status {
        month
        expected {
          amount
          currency
        }
        actual {
          amount
          currency
        }
        gap {
          amount
          currency
        }
      }
    }
    netWorthHistory(period: ONE_MONTH, points: 2) {
      netWorth {
        amount
        currency
      }
    }
    cashFlow(from: $from, months: 12) {
      currency
      months {
        expectedIn {
          amount
        }
        expectedOut {
          amount
        }
      }
    }
  }
`);

/** The plans, plus what a new one would start from: the net worth now and the schedules' monthly surplus. */
export const load: PageLoad = async ({ fetch, url }) => {
  const now = new Date();
  const data = await query(
    PlansQuery,
    { from: new Date(Date.UTC(now.getUTCFullYear(), now.getUTCMonth(), 1)) },
    fetch,
  );
  const months = data.cashFlow.months;
  const surplus = months.length
    ? Math.round(
        months.reduce(
          (sum, m) => sum + m.expectedIn.amount - m.expectedOut.amount,
          0,
        ) / months.length,
      )
    : 0;
  const netWorth = data.netWorthHistory[data.netWorthHistory.length - 1]
    ?.netWorth ?? { amount: 0, currency: data.cashFlow.currency };
  const requested = url.searchParams.get("plan");
  const plan =
    data.plans.find((p) => p.id === requested) ?? data.plans[0] ?? null;
  return {
    plans: data.plans,
    plan,
    netWorth,
    surplus: { amount: surplus, currency: data.cashFlow.currency },
  };
};
