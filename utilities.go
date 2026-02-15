package fap

import "math"

// Distance calculates the great-circle distance in kilometers between two
// points specified in decimal degrees.
func Distance(lat0, lon0, lat1, lon1 float64) float64 {
	lat0r := lat0 * math.Pi / 180.0
	lon0r := lon0 * math.Pi / 180.0
	lat1r := lat1 * math.Pi / 180.0
	lon1r := lon1 * math.Pi / 180.0

	dlon := lon1r - lon0r
	dlat := lat1r - lat0r

	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat0r)*math.Cos(lat1r)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return 6366.71 * c // Earth radius in km
}

// Direction calculates the initial bearing in degrees from point 0 to point 1,
// with both points specified in decimal degrees.
func Direction(lat0, lon0, lat1, lon1 float64) float64 {
	lat0r := lat0 * math.Pi / 180.0
	lon0r := lon0 * math.Pi / 180.0
	lat1r := lat1 * math.Pi / 180.0
	lon1r := lon1 * math.Pi / 180.0

	dlon := lon1r - lon0r

	direction := math.Atan2(
		math.Sin(dlon)*math.Cos(lat1r),
		math.Cos(lat0r)*math.Sin(lat1r)-math.Sin(lat0r)*math.Cos(lat1r)*math.Cos(dlon),
	) * 180.0 / math.Pi

	if direction < 0 {
		direction += 360.0
	}

	return direction
}
