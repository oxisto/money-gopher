-- +goose Up

-- Foreign-currency bond confirmations (US Treasuries, GBP/USD Eurobonds,
-- ...) quote Kurswert/Stückzinsen in the security's own trading currency,
-- convert at a stated rate ("umger. zum Devisenkurs ... (GBP = 0,8519)"),
-- and only then add EUR-denominated fees to reach the EUR Endbetrag. Until
-- now that conversion step was silently ignored: Price/Units/Kurswert were
-- computed correctly in the security's currency, but treated as if they
-- were already the settlement currency.
--
-- price_currency is the currency Price/Units are actually denominated in;
-- NULL means "same as currency" (the common case — no FX involved).
-- fx_rate is 1 unit of currency (settlement) expressed in price_currency,
-- exactly as printed on the document (e.g. 0.8519 for "GBP = 0,8519",
-- meaning 1 EUR = 0.8519 GBP) — dividing a price_currency amount by
-- fx_rate converts it to settlement currency.
ALTER TABLE transactions ADD COLUMN price_currency TEXT;
ALTER TABLE transactions ADD COLUMN fx_rate DECIMAL(19,4);
ALTER TABLE staged_transactions ADD COLUMN price_currency TEXT;
ALTER TABLE staged_transactions ADD COLUMN fx_rate DECIMAL(19,4);

-- +goose Down

-- Not reversible: lightsql has no DROP COLUMN.
