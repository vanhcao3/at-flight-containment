package service

import (
	"fmt"
	"math"
)

func rad(x float64) float64 { return x * math.Pi / 180 }
func latLonToECEF(lat, lon, alt float64) (x, y, z float64) {
	latR := rad(lat)
	lonR := rad(lon)

	a := 6378137.0
	e2 := 6.69437999014e-3

	N := a / math.Sqrt(1-math.Sin(latR)*math.Sin(latR)*e2)

	x = (N + alt) * math.Cos(latR) * math.Cos(lonR)
	y = (N + alt) * math.Cos(latR) * math.Sin(lonR)
	z = (N*(1-e2) + alt) * math.Sin(latR)

	return
}

func ecefToENU(x, y, z float64, lat0, lon0, x0, y0, z0 float64) (e, n, u float64) {
	latR := rad(lat0)
	lonR := rad(lon0)

	dx := x - x0
	dy := y - y0
	dz := z - z0

	e = -math.Sin(lonR)*dx + math.Cos(lonR)*dy
	n = -math.Sin(latR)*math.Cos(lonR)*dx -
		math.Sin(latR)*math.Sin(lonR)*dy +
		math.Cos(latR)*dz
	u = math.Cos(latR)*math.Cos(lonR)*dx +
		math.Cos(latR)*math.Sin(lonR)*dy +
		math.Sin(latR)*dz

	return
}

func latLonAltToENU(lat, lon, alt, refLat, refLon, refAlt float64) (e, n, u float64) {
	x, y, z := latLonToECEF(lat, lon, alt)
	x0, y0, z0 := latLonToECEF(refLat, refLon, refAlt)
	return ecefToENU(x, y, z, refLat, refLon, x0, y0, z0)
}

type Vec struct {
	x, y, z float64
}

func (a Vec) Sub(b Vec) Vec { return Vec{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a Vec) Dot(b Vec) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func (a Vec) Norm() float64 {
	return math.Sqrt(a.Dot(a))
}

func distancePointToSegment(P, A, B Vec) float64 {
	AB := B.Sub(A)
	AP := P.Sub(A)

	t := AP.Dot(AB) / AB.Dot(AB)

	if t <= 0 {
		return AP.Norm()
	} else if t >= 1 {
		return P.Sub(B).Norm()
	}

	closest := Vec{
		A.x + t*AB.x,
		A.y + t*AB.y,
		A.z + t*AB.z,
	}
	return P.Sub(closest).Norm()
}

func closestPointOnSegment(P, A, B Vec) Vec {
	AB := B.Sub(A)
	AP := P.Sub(A)
	denom := AB.Dot(AB)
	if denom == 0 {
		return A
	}
	t := AP.Dot(AB) / denom
	switch {
	case t <= 0:
		return A
	case t >= 1:
		return B
	default:
		return Vec{
			A.x + t*AB.x,
			A.y + t*AB.y,
			A.z + t*AB.z,
		}
	}
}

func closestPointOnPath(P Vec, path []Vec) (Vec, bool) {
	if len(path) < 2 {
		return Vec{}, false
	}
	minDist := math.MaxFloat64
	closest := path[0]
	for i := 0; i < len(path)-1; i++ {
		point := closestPointOnSegment(P, path[i], path[i+1])
		dist := P.Sub(point).Norm()
		if dist < minDist {
			minDist = dist
			closest = point
		}
	}
	return closest, true
}

func compute3DDeviation(drone Vec, path []Vec) float64 {
	minDist := math.MaxFloat64
	for i := 0; i < len(path)-1; i++ {
		d := distancePointToSegment(drone, path[i], path[i+1])
		if d < minDist {
			minDist = d
		}
	}
	return minDist
}

func (ms *MainService) CheckFlightContainment(droneLat, droneLon, droneAlt float64) bool {
	if ms == nil || ms.SvcConfig == nil {
		return false
	}
	settings := ms.SvcConfig.FlightContainment
	if len(settings.Waypoints) < 2 {
		return false
	}
	latThreshold := settings.LatDeviationM
	lonThreshold := settings.LonDeviationM
	altThreshold := settings.AltDeviationM
	horizontalThreshold := math.Max(latThreshold, lonThreshold)
	if horizontalThreshold <= 0 || altThreshold <= 0 {
		return false
	}
	ref := settings.Waypoints[0]
	path := make([]Vec, 0, len(settings.Waypoints))
	for _, w := range settings.Waypoints {
		e, n, u := latLonAltToENU(w.Latitude, w.Longitude, w.Altitude, ref.Latitude, ref.Longitude, ref.Altitude)
		path = append(path, Vec{e, n, u})
	}
	if len(path) < 2 {
		return false
	}
	de, dn, du := latLonAltToENU(droneLat, droneLon, droneAlt, ref.Latitude, ref.Longitude, ref.Altitude)
	drone := Vec{de, dn, du}
	closest, ok := closestPointOnPath(drone, path)
	if !ok {
		return false
	}
	offset := drone.Sub(closest)
	horizontalDeviation := math.Hypot(offset.x, offset.y)
	altDeviation := math.Abs(offset.z)

	fmt.Printf("Drone deviation from path centerline (horizontal: %.3f m, alt: %.3f m)\n", horizontalDeviation, altDeviation)

	if horizontalDeviation > horizontalThreshold || altDeviation > altThreshold {
		fmt.Println("WARNING: Drone is OUTSIDE flight containment cuboid!")
		return true
	}

	fmt.Println("Drone is inside containment cuboid.")
	return false
}
