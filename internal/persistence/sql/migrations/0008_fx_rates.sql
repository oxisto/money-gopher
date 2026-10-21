-- +goose Up

-- Quotes of a listing in one currency (a US Treasury quoted in USD) have to
-- be valued in the currency its portfolio settles in (EUR). The rates come
-- from the ECB's daily reference rates, which are all against the euro:
-- rate is how many units of currency one euro buys on that day, so an
-- amount in currency divided by rate is euros, and euros times rate is the
-- currency. Converting between two non-euro currencies goes through the
-- euro. time is the reference date at midnight UTC.
CREATE TABLE fx_rates (
    currency TEXT      NOT NULL,
    time     TIMESTAMP NOT NULL,
    rate     DECIMAL(19,8) NOT NULL,
    PRIMARY KEY (currency, time)
);

-- +goose Down

DROP TABLE fx_rates;
