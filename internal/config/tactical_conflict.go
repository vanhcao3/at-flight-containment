package config

type TacticalConflictConfig struct {
	SphereRadiusM          float64 `mapstructure:"sphere_radius_m"`
	WarningDistanceM       float64 `mapstructure:"warning_distance_m"`
	DangerDistanceM        float64 `mapstructure:"danger_distance_m"`
	NearCollisionDistanceM float64 `mapstructure:"near_collision_distance_m"`
}
