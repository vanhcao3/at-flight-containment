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

func (a Vec) Add(b Vec) Vec     { return Vec{a.x + b.x, a.y + b.y, a.z + b.z} }
func (a Vec) Sub(b Vec) Vec     { return Vec{a.x - b.x, a.y - b.y, a.z - b.z} }
func (a Vec) Mul(s float64) Vec { return Vec{a.x * s, a.y * s, a.z * s} }
func (a Vec) Dot(b Vec) float64 {
	return a.x*b.x + a.y*b.y + a.z*b.z
}

func (a Vec) Norm() float64 {
	return math.Sqrt(a.Dot(a))
}

func (a Vec) Normalize() Vec {
	n := a.Norm()
	if n == 0 {
		return Vec{}
	}
	return a.Mul(1 / n)
}

func (a Vec) Cross(b Vec) Vec {
	return Vec{
		x: a.y*b.z - a.z*b.y,
		y: a.z*b.x - a.x*b.z,
		z: a.x*b.y - a.y*b.x,
	}
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

func closestPointOnPath(P Vec, path []Vec) (Vec, int, bool) {
	if len(path) < 2 {
		return Vec{}, -1, false
	}
	minDist := math.MaxFloat64
	closest := path[0]
	segment := -1
	for i := 0; i < len(path)-1; i++ {
		point := closestPointOnSegment(P, path[i], path[i+1])
		dist := P.Sub(point).Norm()
		if dist < minDist {
			minDist = dist
			closest = point
			segment = i
		}
	}
	return closest, segment, true
}

type containmentEvaluation struct {
	horizontalExceeded  bool
	verticalExceeded    bool
	horizontalDeviation float64
	verticalDeviation   float64
}

func (ms *MainService) CheckFlightContainment(droneLat, droneLon, droneAlt float64) bool {
	eval, ok := ms.evaluateFlightContainment(droneLat, droneLon, droneAlt)
	if !ok {
		return false
	}
	return eval.horizontalExceeded || eval.verticalExceeded
}

func (ms *MainService) evaluateFlightContainment(droneLat, droneLon, droneAlt float64) (containmentEvaluation, bool) {
	result := containmentEvaluation{}
	if ms == nil || ms.SvcConfig == nil {
		return result, false
	}
	settings := ms.SvcConfig.FlightContainment
	if len(settings.Waypoints) < 2 {
		return result, false
	}
	altThreshold := settings.AltDeviationM
	horizontalThreshold := settings.HorizontalDeviationM
	if horizontalThreshold <= 0 || altThreshold <= 0 {
		return result, false
	}
	path := make([]Vec, 0, len(settings.Waypoints))
	for _, w := range settings.Waypoints {
		x, y, z := latLonToECEF(w.Latitude, w.Longitude, w.Altitude)
		path = append(path, Vec{x, y, z})
	}
	if len(path) < 2 {
		return result, false
	}
	dx, dy, dz := latLonToECEF(droneLat, droneLon, droneAlt)
	drone := Vec{dx, dy, dz}
	closest, segIdx, ok := closestPointOnPath(drone, path)
	if !ok {
		return result, false
	}
	offset := drone.Sub(closest)
	up := closest.Normalize()
	vertical := offset.Dot(up)
	horizontalVec := offset.Sub(up.Mul(vertical))
	horizontalMag := horizontalVec.Norm()

	var signedHorizontal float64
	if horizontalMag > 0 && segIdx >= 0 {
		segmentDir := path[segIdx+1].Sub(path[segIdx]).Normalize()
		right := segmentDir.Cross(up).Normalize()
		if right.Norm() == 0 {
			signedHorizontal = horizontalMag
		} else {
			if horizontalVec.Dot(right) >= 0 {
				signedHorizontal = horizontalMag
			} else {
				signedHorizontal = -horizontalMag
			}
		}
	} else {
		signedHorizontal = 0
	}

	result.horizontalDeviation = signedHorizontal
	result.verticalDeviation = vertical
	result.horizontalExceeded = math.Abs(signedHorizontal) > horizontalThreshold
	result.verticalExceeded = math.Abs(vertical) > altThreshold

	fmt.Printf("Drone deviation from path centerline (horizontal: %.3f m, alt: %.3f m)\n", math.Abs(signedHorizontal), math.Abs(vertical))

	return result, true
}
