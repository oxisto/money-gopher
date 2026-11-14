package plan

import (
	"math"
	"testing"
	"time"
)

var start = time.Date(2026, time.November, 1, 0, 0, 0, 0, time.UTC)

func TestProjectPoints(t *testing.T) {
	got := Project(Input{Start: start, Months: 24, Buckets: []Bucket{{Kind: "CASH", Value: 100, Weight: 1}}})
	if len(got) != 25 {
		t.Fatalf("points = %d, want 25", len(got))
	}
	if !got[0].Month.Equal(start) || !got[24].Month.Equal(start.AddDate(2, 0, 0)) {
		t.Errorf("months %v … %v", got[0].Month, got[24].Month)
	}
}

func TestZeroReturnIsStartPlusContributions(t *testing.T) {
	got := Project(Input{Start: start, Months: 12, Contribution: 50_000,
		Buckets: []Bucket{{Kind: "CASH", Value: 1_000_000, Weight: 1}}})
	last := got[12]
	if last.Expected != 1_600_000 || last.Contributed != 600_000 {
		t.Errorf("after a year: expected %d contributed %d, want 1,600,000 / 600,000", last.Expected, last.Contributed)
	}
	if got[0].Expected != 1_000_000 || got[0].Contributed != 0 {
		t.Errorf("first point %+v, want the start untouched", got[0])
	}
}

func TestCompoundsMonthly(t *testing.T) {
	got := Project(Input{Start: start, Months: 12,
		Buckets: []Bucket{{Kind: "ETF", Value: 1_000_000, Weight: 1, Return: 0.12}}})
	if got[12].Expected != 1_120_000 {
		t.Errorf("12 %% on 10,000 after a year = %d, want 1,120,000", got[12].Expected)
	}
	// Half a year at (1.12)^(1/2).
	want := int64(math.Round(1_000_000 * math.Sqrt(1.12)))
	if got[6].Expected != want {
		t.Errorf("after six months = %d, want %d", got[6].Expected, want)
	}
}

func TestCostLowersThePath(t *testing.T) {
	free := Project(Input{Start: start, Months: 120, Buckets: []Bucket{{Kind: "ETF", Value: 1_000_000, Weight: 1, Return: 0.07}}})
	costly := Project(Input{Start: start, Months: 120, Cost: 0.005, Buckets: []Bucket{{Kind: "ETF", Value: 1_000_000, Weight: 1, Return: 0.07}}})
	if costly[120].Expected >= free[120].Expected {
		t.Errorf("with cost %d, without %d", costly[120].Expected, free[120].Expected)
	}
	if want := int64(math.Round(1_000_000 * math.Pow(1.065, 10))); costly[120].Expected != want {
		t.Errorf("6.5 %% net over ten years = %d, want %d", costly[120].Expected, want)
	}
}

func TestDebtStaysFlat(t *testing.T) {
	got := Project(Input{Start: start, Months: 12, Cost: 0.01, Contribution: 100,
		Buckets: []Bucket{{Kind: "DEBT", Value: -19_000_000}, {Kind: "CASH", Value: 0, Weight: 1}}})
	if got[12].Expected != -19_000_000+1_200 {
		t.Errorf("debt plus a year of cash = %d, want -18,998,800", got[12].Expected)
	}
}

func TestContributionSplitByWeight(t *testing.T) {
	got := Project(Input{Start: start, Months: 1, Contribution: 1_000,
		Buckets: []Bucket{{Kind: "ETF", Weight: 0.8, Return: 0.06}, {Kind: "SAVINGS", Weight: 0.2, Return: 0.02}}})
	if got[1].Expected != 1_000 {
		t.Errorf("first contribution lands unchanged: %d", got[1].Expected)
	}
	two := Project(Input{Start: start, Months: 2, Contribution: 1_000,
		Buckets: []Bucket{{Kind: "ETF", Weight: 0.8, Return: 0.06}, {Kind: "SAVINGS", Weight: 0.2, Return: 0.02}}})
	want := 800*math.Pow(1.06, 1.0/12) + 200*math.Pow(1.02, 1.0/12) + 1_000
	if two[2].Expected != int64(math.Round(want)) {
		t.Errorf("second month = %d, want %d", two[2].Expected, int64(math.Round(want)))
	}
}

func TestDefaultSplit(t *testing.T) {
	got := DefaultSplit([]Bucket{
		{Kind: "ETF", Value: 3_000}, {Kind: "CASH", Value: 1_000}, {Kind: "PROPERTY", Value: 25_000_000}, {Kind: "DEBT", Value: -19_000_000},
	})
	if got[0].Weight != 0.75 || got[1].Weight != 0.25 || got[2].Weight != 0 || got[3].Weight != 0 {
		t.Errorf("weights = %v %v %v %v, want 0.75 0.25 0 0", got[0].Weight, got[1].Weight, got[2].Weight, got[3].Weight)
	}

	declared := DefaultSplit([]Bucket{{Kind: "ETF", Value: 3_000, Weight: 1}, {Kind: "SAVINGS", Value: 1_000, Weight: 3}})
	if declared[0].Weight != 0.25 || declared[1].Weight != 0.75 {
		t.Errorf("declared weights normalised = %v %v, want 0.25 0.75", declared[0].Weight, declared[1].Weight)
	}

	empty := DefaultSplit([]Bucket{{Kind: "PROPERTY", Value: 25_000_000}, {Kind: "CASH"}})
	if empty[1].Weight != 1 {
		t.Errorf("nothing liquid held: CASH weight = %v, want 1", empty[1].Weight)
	}
	added := DefaultSplit([]Bucket{{Kind: "PROPERTY", Value: 25_000_000}})
	if len(added) != 2 || added[1].Kind != "CASH" || added[1].Weight != 1 {
		t.Errorf("no CASH bucket: %+v, want one added with weight 1", added)
	}
}

func TestFan(t *testing.T) {
	in := Input{Start: start, Months: 120, Contribution: 50_000, Paths: 2_000, Seed: 7,
		Buckets: []Bucket{
			{Kind: "ETF", Value: 5_000_000, Weight: 0.8, Return: 0.06, Volatility: 0.15},
			{Kind: "SAVINGS", Value: 1_000_000, Weight: 0.2, Return: 0.02},
			{Kind: "DEBT", Value: -3_000_000},
		}}
	got := Project(in)
	for i, p := range got {
		if !(p.P10 <= p.P25 && p.P25 <= p.P50 && p.P50 <= p.P75 && p.P75 <= p.P90) {
			t.Fatalf("month %d: percentiles out of order %+v", i, p)
		}
	}
	if p := got[0]; p.P10 != p.Expected || p.P90 != p.Expected {
		t.Errorf("the start has no spread: %+v", p)
	}
	last := got[120]
	if last.P10 == last.P90 {
		t.Errorf("no spread after ten years at 15 %% volatility: %+v", last)
	}
	// The median of a geometric Brownian motion is the drift path, which
	// is the expected path here by construction.
	if ratio := float64(last.P50) / float64(last.Expected); ratio < 0.93 || ratio > 1.07 {
		t.Errorf("median %d vs expected %d (ratio %.3f)", last.P50, last.Expected, ratio)
	}

	again := Project(in)
	if again[120] != got[120] {
		t.Errorf("same seed, different fan: %+v vs %+v", again[120], got[120])
	}
	in.Seed = 8
	if other := Project(in); other[120].P10 == got[120].P10 && other[120].P90 == got[120].P90 {
		t.Errorf("another seed gave the same fan")
	}
}

func TestFanWithoutVolatilityIsTheExpectedPath(t *testing.T) {
	got := Project(Input{Start: start, Months: 36, Contribution: 1_000, Paths: 50, Seed: 1,
		Buckets: []Bucket{{Kind: "SAVINGS", Value: 100_000, Weight: 1, Return: 0.03}}})
	for i, p := range got {
		if p.P10 != p.Expected || p.P90 != p.Expected {
			t.Fatalf("month %d: %+v, want every percentile on the expected path", i, p)
		}
	}
}

func TestPercentile(t *testing.T) {
	sorted := []float64{10, 20, 30, 40, 50}
	for _, tc := range []struct{ q, want float64 }{{0, 10}, {0.5, 30}, {1, 50}, {0.25, 20}, {0.1, 14}} {
		if got := percentile(sorted, tc.q); got != tc.want {
			t.Errorf("percentile(%v) = %v, want %v", tc.q, got, tc.want)
		}
	}
	if percentile(nil, 0.5) != 0 {
		t.Errorf("empty input")
	}
}
