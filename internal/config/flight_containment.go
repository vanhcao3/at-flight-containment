package config

type FlightContainmentConfig struct {
	Radius    float64                           `mapstructure:"radius"`
	Waypoints []FlightContainmentWaypointConfig `mapstructure:"waypoints"`
}

type FlightContainmentWaypointConfig struct {
	Latitude  float64 `mapstructure:"lat"`
	Longitude float64 `mapstructure:"lon"`
	Altitude  float64 `mapstructure:"alt"`
}
