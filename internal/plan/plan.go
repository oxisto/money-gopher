// Package plan projects a net worth into the future: a deterministic path
// from per-bucket return assumptions, and a fan of simulated paths around
// it from their volatilities.
//
// A bucket is one kind of wealth (listed shares, the house, cash, a debt)
// with a value today, a share of every future contribution, and an
// expected annual return and volatility. Buckets are treated as
// uncorrelated, and a bucket with no return (a debt whose interest is paid
// from the cash flow, say) only receives contributions. All amounts are
// integer minor units.
package plan

import (
	"math"
	"math/rand/v2"
	"sort"
	"time"
)

// Bucket is one kind of wealth as the plan sees it.
type Bucket struct {
	Kind string
	// Value is what the bucket holds at the start, negative for a debt.
	Value int64
	// Weight is the share of each contribution the bucket receives; the
	// weights of all buckets sum to one.
	Weight float64
	// Return is the compound annual growth rate expected of the bucket
	// (0.07 for 7 %); Volatility the annual standard deviation of its
	// log return.
	Return, Volatility float64
}

// Input is everything a projection needs.
type Input struct {
	// Start is the first month; the projection's first point sits there.
	Start time.Time
	// Months is the horizon; the projection has Months+1 points.
	Months int
	Buckets []Bucket
	// Contribution is added at the start of every month after the first,
	// split among the buckets by weight.
	Contribution int64
	// Cost is an annual fraction taken off every growing bucket's return.
	Cost float64
	// Paths is the number of simulated paths behind the percentiles; zero
	// leaves the percentiles at the expected path.
	Paths int
	Seed  uint64
}

// Point is the projection at the start of one month.
type Point struct {
	Month time.Time
	// Expected is the deterministic path: every bucket grown at its return
	// net of cost, plus the contributions so far.
	Expected int64
	// Contributed is the sum of the contributions up to this month.
	Contributed int64
	// The percentiles of the simulated paths.
	P10, P25, P50, P75, P90 int64
}

// Liquid reports whether a kind of wealth is one that a contribution
// can flow into on its own: listed holdings, cash and savings.
func Liquid(kind string) bool {
	switch kind {
	case "STOCK", "ETF", "FUND", "BOND", "CASH", "SAVINGS":
		return true
	}
	return false
}

// DefaultSplit fills in the weights of the buckets that have none,
// following the current allocation among the liquid buckets with positive
// values; when none is held, everything goes to CASH (added if missing).
// Buckets with a weight keep it, and the result's weights sum to one.
func DefaultSplit(buckets []Bucket) []Bucket {
	out := make([]Bucket, len(buckets))
	copy(out, buckets)
	var declared float64
	anyDeclared := false
	for _, b := range out {
		if b.Weight > 0 {
			declared += b.Weight
			anyDeclared = true
		}
	}
	if anyDeclared {
		for i := range out {
			out[i].Weight /= declared
		}
		return out
	}
	var liquid float64
	for _, b := range out {
		if Liquid(b.Kind) && b.Value > 0 {
			liquid += float64(b.Value)
		}
	}
	if liquid > 0 {
		for i := range out {
			if Liquid(out[i].Kind) && out[i].Value > 0 {
				out[i].Weight = float64(out[i].Value) / liquid
			}
		}
		return out
	}
	for i := range out {
		if out[i].Kind == "CASH" {
			out[i].Weight = 1
			return out
		}
	}
	return append(out, Bucket{Kind: "CASH", Weight: 1})
}

// Project runs the deterministic path and the simulation.
func Project(in Input) []Point {
	if in.Months < 0 {
		in.Months = 0
	}
	n := in.Months + 1
	points := make([]Point, n)
	for i := range points {
		points[i].Month = in.Start.AddDate(0, i, 0)
	}

	// The expected path: every bucket compounds monthly at its net return.
	values := make([]float64, len(in.Buckets))
	growth := make([]float64, len(in.Buckets))
	for i, b := range in.Buckets {
		values[i] = float64(b.Value)
		growth[i] = monthlyGrowth(b, in.Cost)
	}
	var contributed int64
	for m := 0; m < n; m++ {
		if m > 0 {
			contributed += in.Contribution
			for i, b := range in.Buckets {
				values[i] = values[i]*growth[i] + float64(in.Contribution)*b.Weight
			}
		}
		var total float64
		for _, v := range values {
			total += v
		}
		points[m].Expected = round(total)
		points[m].Contributed = contributed
		points[m].P10, points[m].P25, points[m].P50, points[m].P75, points[m].P90 = points[m].Expected, points[m].Expected, points[m].Expected, points[m].Expected, points[m].Expected
	}

	if in.Paths <= 0 {
		return points
	}
	simulate(in, points)
	return points
}

// monthlyGrowth is the factor a bucket's value is multiplied by per month
// on the expected path: its annual return net of cost, compounded monthly.
// A bucket without a return (a debt, an idle kind) stays flat; the cost
// does not eat it.
func monthlyGrowth(b Bucket, cost float64) float64 {
	if b.Return == 0 {
		return 1
	}
	annual := 1 + b.Return - cost
	if annual <= 0 {
		return 0
	}
	return math.Pow(annual, 1.0/12)
}

// simulate draws Paths geometric Brownian paths per bucket and writes the
// percentiles of their monthly sums into the points.
func simulate(in Input, points []Point) {
	n := len(points)
	rng := rand.New(rand.NewPCG(in.Seed, in.Seed))
	sums := make([][]float64, n)
	for m := range sums {
		sums[m] = make([]float64, in.Paths)
	}
	drift := make([]float64, len(in.Buckets))
	sigma := make([]float64, len(in.Buckets))
	for i, b := range in.Buckets {
		if b.Return == 0 {
			continue
		}
		// Return is a compound annual rate, so the simulated median
		// follows the expected path and the mean sits above it.
		drift[i] = math.Log1p(b.Return-in.Cost) / 12
		sigma[i] = b.Volatility / math.Sqrt(12)
	}
	values := make([]float64, len(in.Buckets))
	for p := 0; p < in.Paths; p++ {
		for i, b := range in.Buckets {
			values[i] = float64(b.Value)
		}
		for m := 0; m < n; m++ {
			if m > 0 {
				for i, b := range in.Buckets {
					if b.Return != 0 {
						z := rng.NormFloat64()
						values[i] *= math.Exp(drift[i] + sigma[i]*z)
					}
					values[i] += float64(in.Contribution) * b.Weight
				}
			}
			var total float64
			for _, v := range values {
				total += v
			}
			sums[m][p] = total
		}
	}
	for m := range points {
		sort.Float64s(sums[m])
		points[m].P10 = round(percentile(sums[m], 0.10))
		points[m].P25 = round(percentile(sums[m], 0.25))
		points[m].P50 = round(percentile(sums[m], 0.50))
		points[m].P75 = round(percentile(sums[m], 0.75))
		points[m].P90 = round(percentile(sums[m], 0.90))
	}
}

// percentile reads a quantile off sorted values with linear interpolation.
func percentile(sorted []float64, q float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	pos := q * float64(len(sorted)-1)
	lo := int(math.Floor(pos))
	hi := int(math.Ceil(pos))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo] + (sorted[hi]-sorted[lo])*(pos-float64(lo))
}

func round(v float64) int64 {
	return int64(math.Round(v))
}
