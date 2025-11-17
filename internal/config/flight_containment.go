package config

type FlightContainmentConfig struct {
	Radius          float64                           `mapstructure:"radius"`
	Waypoints       []FlightContainmentWaypointConfig `mapstructure:"waypoints"`
	RenotifySeconds float64                           `mapstructure:"renotify_seconds"`
	LatDeviationM   float64                           `mapstructure:"lat_deviation_m"`
	LonDeviationM   float64                           `mapstructure:"lon_deviation_m"`
	AltDeviationM   float64                           `mapstructure:"alt_deviation_m"`
}

type FlightContainmentWaypointConfig struct {
	Latitude  float64 `mapstructure:"lat"`
	Longitude float64 `mapstructure:"lon"`
	Altitude  float64 `mapstructure:"alt"`
}
