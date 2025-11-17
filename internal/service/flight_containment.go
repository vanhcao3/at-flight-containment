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

	radius := 5.0 // meters

	// Intended waypoint coordinates in decimal degrees + altitude
	waypoints := [][]float64{
		{21.002694, 105.537611, 40.0}, // P1
		{21.001444, 105.538111, 40.0}, // P2
		{21.000500, 105.535889, 40.0}, // P3
		{21.001944, 105.535222, 40.0}, // P4
	}

	// Reference point for ENU origin
	refLat := waypoints[0][0]
	refLon := waypoints[0][1]
	refAlt := waypoints[0][2]

	// Convert intended path to ENU (meters)
	path := []Vec{}
	for _, w := range waypoints {
		e, n, u := latLonAltToENU(w[0], w[1], w[2], refLat, refLon, refAlt)
		path = append(path, Vec{e, n, u})
	}

	de, dn, du := latLonAltToENU(droneLat, droneLon, droneAlt, refLat, refLon, refAlt)
	drone := Vec{de, dn, du}

	dev := compute3DDeviation(drone, path)

	fmt.Printf("Drone deviation from path centerline: %.3f m\n", dev)

	if dev > radius {
		fmt.Println("WARNING: Drone is OUTSIDE cylindrical flight containment!")
		return true
	} else {
		fmt.Println("Drone is inside containment corridor.")
		return false
	}
}
