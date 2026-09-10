package profile

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

// earthRadiusKM is the mean radius used by the Haversine formula.
const earthRadiusKM = 6371

// DistanceKM returns the great-circle distance between two points in
// kilometres. The profile search query has its own copy of this formula in
// SQL; that one filters rows, this one labels them for the response.
func DistanceKM(lat1, lon1, lat2, lon2 float64) float64 {
	dLat := degreesToRadians(lat2 - lat1)
	dLon := degreesToRadians(lon2 - lon1)

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(degreesToRadians(lat1))*math.Cos(degreesToRadians(lat2))*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	return earthRadiusKM * 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

func degreesToRadians(d float64) float64 {
	return d * math.Pi / 180
}

// Bander turns an exact distance into a coarse band such as "5–10 km".
// Bucketing is what protects the position: an exact distance, sampled from a
// few places, pins a user down by trilateration, while a band only ever
// narrows them to the width of the bucket.
type Bander struct {
	thresholds []float64
}

// NewBander validates the thresholds so a bad deployment config fails at
// startup instead of silently mislabelling distances at runtime. Values are
// kilometres, ascending, e.g. {5, 10, 25, 50}.
func NewBander(thresholds []float64) (*Bander, error) {
	if len(thresholds) == 0 {
		return nil, fmt.Errorf("distance bands: at least one threshold is required")
	}
	for i, t := range thresholds {
		if t <= 0 {
			return nil, fmt.Errorf("distance bands: threshold %v must be positive", t)
		}
		if i > 0 && t <= thresholds[i-1] {
			return nil, fmt.Errorf("distance bands: thresholds must ascend, got %v after %v", t, thresholds[i-1])
		}
	}
	return &Bander{thresholds: thresholds}, nil
}

// ParseBands converts the configured threshold strings into a Bander, so the
// band granularity is deployment configuration rather than a constant.
func ParseBands(values []string) (*Bander, error) {
	thresholds := make([]float64, 0, len(values))
	for _, v := range values {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("distance bands: %q is not a number", v)
		}
		thresholds = append(thresholds, f)
	}
	sort.Float64s(thresholds)
	return NewBander(thresholds)
}

// Band labels a distance. Labels are generated from the thresholds, so
// changing the configuration changes the buckets with no code change:
// {5,10,25} yields "< 5 km", "5–10 km", "10–25 km", "25+ km".
func (b *Bander) Band(km float64) string {
	for i, t := range b.thresholds {
		if km < t {
			if i == 0 {
				return fmt.Sprintf("< %s km", formatKM(t))
			}
			return fmt.Sprintf("%s–%s km", formatKM(b.thresholds[i-1]), formatKM(t))
		}
	}
	return fmt.Sprintf("%s+ km", formatKM(b.thresholds[len(b.thresholds)-1]))
}

// formatKM keeps whole numbers whole ("5" not "5.0") while still allowing
// fractional thresholds for dense cities.
func formatKM(v float64) string {
	return strconv.FormatFloat(v, 'f', -1, 64)
}
