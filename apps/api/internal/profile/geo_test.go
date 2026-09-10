package profile_test

import (
	"testing"

	"github.com/kisssonik/hearts/internal/profile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistanceKM(t *testing.T) {
	// London centre to Heathrow is about 22 km.
	km := profile.DistanceKM(51.5074, -0.1278, 51.4700, -0.4543)
	assert.InDelta(t, 22, km, 1.5)

	// The same point is zero distance from itself.
	assert.InDelta(t, 0, profile.DistanceKM(51.5074, -0.1278, 51.5074, -0.1278), 0.0001)
}

func TestBand_LabelsAreGeneratedFromThresholds(t *testing.T) {
	bander, err := profile.ParseBands([]string{"5", "10", "25", "50"})
	require.NoError(t, err)

	assert.Equal(t, "< 5 km", bander.Band(0))
	assert.Equal(t, "< 5 km", bander.Band(4.9))
	assert.Equal(t, "5–10 km", bander.Band(5))
	assert.Equal(t, "5–10 km", bander.Band(9.9))
	assert.Equal(t, "10–25 km", bander.Band(10))
	assert.Equal(t, "25–50 km", bander.Band(25))
	assert.Equal(t, "50+ km", bander.Band(50))
	assert.Equal(t, "50+ km", bander.Band(4000))
}

// A different configuration must produce different buckets with no code
// change — that is the whole point of the thresholds being config.
func TestBand_DifferentConfigurationDifferentBuckets(t *testing.T) {
	bander, err := profile.ParseBands([]string{"2", "10", "50"})
	require.NoError(t, err)

	assert.Equal(t, "< 2 km", bander.Band(1))
	assert.Equal(t, "2–10 km", bander.Band(5))
	assert.Equal(t, "10–50 km", bander.Band(22))
	assert.Equal(t, "50+ km", bander.Band(60))

	single, err := profile.ParseBands([]string{"25"})
	require.NoError(t, err)
	assert.Equal(t, "< 25 km", single.Band(10))
	assert.Equal(t, "25+ km", single.Band(30))
}

func TestNewBander_RejectsBadConfiguration(t *testing.T) {
	_, err := profile.NewBander(nil)
	assert.Error(t, err)

	_, err = profile.NewBander([]float64{0})
	assert.Error(t, err)

	_, err = profile.NewBander([]float64{-5})
	assert.Error(t, err)

	// Duplicates would produce an empty bucket like "5–5 km"
	_, err = profile.NewBander([]float64{5, 5})
	assert.Error(t, err)

	_, err = profile.ParseBands([]string{"5", "abc"})
	assert.Error(t, err)
}

// Thresholds given out of order are sorted rather than rejected, so operators
// are not tripped up by the ordering of an env var.
func TestParseBands_SortsInput(t *testing.T) {
	bander, err := profile.ParseBands([]string{"25", "5", "10"})
	require.NoError(t, err)
	assert.Equal(t, "< 5 km", bander.Band(1))
	assert.Equal(t, "10–25 km", bander.Band(20))
}
