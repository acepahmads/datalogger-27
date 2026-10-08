package aggregation

import (
	"fmt"
	"time"
)

// TimeBucket holds deterministic boundary timestamps and period identifier
type TimeBucket struct {
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Identifier  string    `json:"identifier"` // YYYYMMDDHHmmss based on PeriodStart in bucket timezone
	Timezone    string    `json:"timezone"`
}

// ResolveLocation returns a valid time.Location for a timezone string, falling back to time.UTC
func ResolveLocation(tz string) *time.Location {
	if tz == "" || tz == "UTC" {
		return time.UTC
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return loc
}

// FormatIdentifier generates the deterministic identifier YYYYMMDDHHmmss from periodStart in given location
func FormatIdentifier(periodStart time.Time, loc *time.Location) string {
	if loc == nil {
		loc = time.UTC
	}
	localStart := periodStart.In(loc)
	return localStart.Format("20060102150405")
}

// ParseIdentifier parses an identifier in format YYYYMMDDHHmmss back into time.Time for the given timezone
func ParseIdentifier(identifier string, loc *time.Location) (time.Time, error) {
	if loc == nil {
		loc = time.UTC
	}
	if len(identifier) != 14 {
		return time.Time{}, fmt.Errorf("invalid identifier length: expected 14 digits (YYYYMMDDHHmmss), got %d", len(identifier))
	}
	t, err := time.ParseInLocation("20060102150405", identifier, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid identifier format: %w", err)
	}
	return t, nil
}

// CalculateBucket computes the deterministic time bucket containing the given timestamp
func CalculateBucket(t time.Time, intervalSeconds int, tz string) TimeBucket {
	loc := ResolveLocation(tz)
	if intervalSeconds <= 0 {
		intervalSeconds = 300 // Default 5 minutes
	}

	localT := t.In(loc)

	// Determine period start by aligning to interval from day start or epoch
	var startLocal time.Time
	if 86400%intervalSeconds == 0 {
		// Interval divides a day cleanly (e.g. 2m=120s, 5m=300s, 10m=600s, 15m=900s, 30m=1800s, 60m=3600s)
		dayStart := time.Date(localT.Year(), localT.Month(), localT.Day(), 0, 0, 0, 0, loc)
		secIntoDay := int(localT.Sub(dayStart).Seconds())
		bucketIndex := secIntoDay / intervalSeconds
		startLocal = dayStart.Add(time.Duration(bucketIndex*intervalSeconds) * time.Second)
	} else {
		// Generic custom duration from Unix epoch
		unixSec := localT.Unix()
		bucketIndex := unixSec / int64(intervalSeconds)
		startLocal = time.Unix(bucketIndex*int64(intervalSeconds), 0).In(loc)
	}

	endLocal := startLocal.Add(time.Duration(intervalSeconds) * time.Second)
	identifier := startLocal.Format("20060102150405")

	return TimeBucket{
		PeriodStart: startLocal.UTC(),
		PeriodEnd:   endLocal.UTC(),
		Identifier:  identifier,
		Timezone:    loc.String(),
	}
}

// GetCompletedBuckets returns all contiguous buckets between startTime and upTo whose PeriodEnd <= upTo
func GetCompletedBuckets(startTime, upTo time.Time, intervalSeconds, gracePeriodSeconds int, tz string) []TimeBucket {
	if intervalSeconds <= 0 {
		intervalSeconds = 300
	}
	if gracePeriodSeconds < 0 {
		gracePeriodSeconds = 0
	}

	// In real-time aggregation, any bucket whose period has elapsed (PeriodEnd <= upTo)
	// is immediately completed and eligible for rollup evaluation.
	maxAllowedTime := upTo
	if startTime.After(maxAllowedTime) {
		return nil
	}

	var buckets []TimeBucket
	currBucket := CalculateBucket(startTime, intervalSeconds, tz)

	// Guard against infinite loop
	maxBuckets := 1000
	count := 0

	for currBucket.PeriodEnd.Before(maxAllowedTime) || currBucket.PeriodEnd.Equal(maxAllowedTime) {
		buckets = append(buckets, currBucket)
		count++
		if count >= maxBuckets {
			break
		}
		// Next bucket starts at current bucket's period end
		currBucket = CalculateBucket(currBucket.PeriodEnd, intervalSeconds, tz)
	}

	return buckets
}

// GetCurrentActiveBucket returns the currently active in-progress bucket for the given timestamp
func GetCurrentActiveBucket(now time.Time, intervalSeconds int, tz string) TimeBucket {
	return CalculateBucket(now, intervalSeconds, tz)
}
