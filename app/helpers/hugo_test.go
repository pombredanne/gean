package helpers

import (
	"testing"

	"github.com/govenue/assert"
	"github.com/govenue/require"
)

func TestHugoVersion(t *testing.T) {
	assert.Equal(t, "0.15-DEV", hugoVersion(0.15, 0, "-DEV"))
	assert.Equal(t, "0.15.2-DEV", hugoVersion(0.15, 2, "-DEV"))

	v := HugoVersion{Number: 0.21, PatchLevel: 0, Suffix: "-DEV"}

	require.Equal(t, v.ReleaseVersion().String(), "0.21")
	require.Equal(t, "0.21-DEV", v.String())
	require.Equal(t, "0.22", v.Next().String())
	require.Equal(t, "0.20.3", v.NextPatchLevel(3).String())
}

func TestCompareVersions(t *testing.T) {
	require.Equal(t, 0, compareVersions(0.20, 0, 0.20))
	require.Equal(t, 0, compareVersions(0.20, 0, float32(0.20)))
	require.Equal(t, 0, compareVersions(0.20, 0, float64(0.20)))
	require.Equal(t, 1, compareVersions(0.19, 1, 0.20))
	require.Equal(t, 1, compareVersions(0.19, 3, "0.20.2"))
	require.Equal(t, -1, compareVersions(0.19, 1, 0.01))
	require.Equal(t, 1, compareVersions(0, 1, 3))
	require.Equal(t, 1, compareVersions(0, 1, int32(3)))
	require.Equal(t, 1, compareVersions(0, 1, int64(3)))
	require.Equal(t, 0, compareVersions(0.20, 0, "0.20"))
	require.Equal(t, 0, compareVersions(0.20, 1, "0.20.1"))
	require.Equal(t, -1, compareVersions(0.20, 1, "0.20"))
	require.Equal(t, 1, compareVersions(0.20, 0, "0.20.1"))
	require.Equal(t, 1, compareVersions(0.20, 1, "0.20.2"))
	require.Equal(t, 1, compareVersions(0.21, 1, "0.22.1"))
}

func TestParseHugoVersion(t *testing.T) {
	require.Equal(t, "0.25", MustParseHugoVersion("0.25").String())
	require.Equal(t, "0.25.2", MustParseHugoVersion("0.25.2").String())
	require.Equal(t, "0.25-test", MustParseHugoVersion("0.25-test").String())
	_, err := ParseHugoVersion("0.25-DEV")
	require.Error(t, err)
}
