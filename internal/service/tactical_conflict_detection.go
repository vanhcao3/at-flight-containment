package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TacticalConflictLevel string

const (
	TacticalConflictLevelWarning       TacticalConflictLevel = "warning"
	TacticalConflictLevelDanger        TacticalConflictLevel = "danger"
	TacticalConflictLevelNearCollision TacticalConflictLevel = "near_collision"
)

type TacticalConflictNotification struct {
	TrackAID         int32                 `json:"track_a"`
	TrackBID         int32                 `json:"track_b"`
	Level            TacticalConflictLevel `json:"level"`
	SeparationMeters float64               `json:"separation_meters"`
}

func (ms *MainService) StartTacticalConflictMonitor(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = time.Second
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := ms.detectAndNotifyTacticalConflicts(ctx); err != nil && ctx.Err() == nil {
					config.PrintErrorLog(ctx, err, "Tactical conflict monitor failed")
				}
			}
		}
	}()
}

func (ms *MainService) detectAndNotifyTacticalConflicts(ctx context.Context) error {
	conflicts, err := ms.detectTacticalConflicts(ctx)
	if err != nil {
		return err
	}
	if len(conflicts) == 0 {
		return nil
	}
	return ms.Notifier().Publish(EventTacticalConflictDetected, conflicts)
}

func (ms *MainService) detectTacticalConflicts(ctx context.Context) ([]TacticalConflictNotification, error) {
	cfg := ms.SvcConfig.TacticalConflict
	if err := validateTacticalConflictConfig(cfg); err != nil {
		return nil, err
	}

	tracks, err := ms.FindAllInMemObjectTrack(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}

	positions := make([]struct {
		id    int32
		vec   Vec
		valid bool
	}, len(tracks))

	for i, t := range tracks {
		id := t.ObjectTrackID
		if t == nil || t.Position == nil || id == 0 {
			continue
		}
		x, y, z := latLonToECEF(float64(t.Position.Latitude), float64(t.Position.Longitude), float64(t.Position.Altitude))
		positions[i] = struct {
			id    int32
			vec   Vec
			valid bool
		}{
			id:    id,
			vec:   Vec{x: x, y: y, z: z},
			valid: true,
		}
	}

	seen := make(map[string]struct{})
	conflicts := []TacticalConflictNotification{}
	for i := 0; i < len(positions); i++ {
		if !positions[i].valid {
			continue
		}
		for j := i + 1; j < len(positions); j++ {
			if !positions[j].valid || positions[i].id == positions[j].id {
				continue
			}
			dist := positions[i].vec.Sub(positions[j].vec).Norm()
			if dist > 2*cfg.SphereRadiusM {
				continue
			}
			level, ok := determineTacticalConflictLevel(dist, cfg)
			if !ok {
				continue
			}
			key := conflictPairKey(positions[i].id, positions[j].id)
			if _, exists := seen[key]; exists {
				continue
			}
			seen[key] = struct{}{}
			conflicts = append(conflicts, TacticalConflictNotification{
				TrackAID:         positions[i].id,
				TrackBID:         positions[j].id,
				Level:            level,
				SeparationMeters: dist,
			})
		}
	}

	return conflicts, nil
}

func determineTacticalConflictLevel(distance float64, cfg config.TacticalConflictConfig) (TacticalConflictLevel, bool) {
	if cfg.NearCollisionDistanceM > 0 && distance <= cfg.NearCollisionDistanceM {
		return TacticalConflictLevelNearCollision, true
	}
	if cfg.DangerDistanceM > 0 && distance <= cfg.DangerDistanceM {
		return TacticalConflictLevelDanger, true
	}
	if cfg.WarningDistanceM > 0 && distance <= cfg.WarningDistanceM {
		return TacticalConflictLevelWarning, true
	}
	return "", false
}

func validateTacticalConflictConfig(cfg config.TacticalConflictConfig) error {
	if cfg.SphereRadiusM <= 0 {
		return errors.New("tactical conflict detection disabled: sphere radius must be positive")
	}
	if cfg.WarningDistanceM <= 0 || cfg.DangerDistanceM <= 0 || cfg.NearCollisionDistanceM <= 0 {
		return errors.New("tactical conflict detection disabled: distance thresholds must be positive")
	}
	return nil
}

func conflictPairKey(a, b int32) string {
	if a < b {
		return fmt.Sprintf("%d-%d", a, b)
	}
	return fmt.Sprintf("%d-%d", b, a)
}
