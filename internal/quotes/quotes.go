// Package quotes provides pluggable quote providers and the updater that
// periodically fetches quotes for all listings and appends them to the
// quote history.
package quotes

import (
	"context"
	"errors"
	"time"
)

// ErrEmptyResult is returned when a provider responded, but without a quote.
var ErrEmptyResult = errors.New("provider returned an empty result")

// Instrument carries everything a provider might need to identify a listed
// security: the listing's own attributes plus the security's external
// identifiers. Not every provider uses every field — Yahoo goes by ticker,
// ING by ISIN.
type Instrument struct {
	Ticker   string
	Exchange string
	// Currency is the ISO 4217 code the listing's quotes are denominated in.
	Currency string
	ISIN     string
	WKN      string
}

// Quote is one price observation returned by a provider.
type Quote struct {
	// Price in minor units of the listing currency.
	Price int64
	// Time is when the price was observed, as reported by the provider.
	Time time.Time
}

// Provider fetches quotes from one external source.
type Provider interface {
	// Name is the identifier stored in a listing's quote_provider column.
	Name() string
	// LatestQuote returns the most recent quote for the instrument.
	LatestQuote(ctx context.Context, ins Instrument) (Quote, error)
}

// Registry holds the available providers by name.
type Registry struct {
	providers map[string]Provider
}

// NewRegistry builds a registry from the given providers.
func NewRegistry(providers ...Provider) *Registry {
	r := &Registry{providers: make(map[string]Provider, len(providers))}
	for _, p := range providers {
		r.providers[p.Name()] = p
	}
	return r
}

// DefaultRegistry contains all built-in providers.
func DefaultRegistry() *Registry {
	return NewRegistry(NewYahoo(), NewING())
}

// Get returns the provider with the given name.
func (r *Registry) Get(name string) (Provider, bool) {
	p, ok := r.providers[name]
	return p, ok
}

// Names returns the names of all registered providers, for display purposes.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
