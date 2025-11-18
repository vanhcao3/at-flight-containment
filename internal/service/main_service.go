package service

import (
	"context"
	"reflect"
	"strings"
	"sync"
	"time"

	config "172.21.5.249/air-trans/at-drone/internal/config"
	gclient "172.21.5.249/air-trans/at-drone/internal/gapi/client"
	pb "172.21.5.249/air-trans/at-drone/pkg/pb"

	moptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/go-co-op/gocron"
	"github.com/nats-io/nats.go"
	"github.com/qiniu/qmgo"
	"github.com/qiniu/qmgo/options"
	"github.com/rs/zerolog/log"
)

const (
	DRONE = "drone"
	// track_history = "track_history"
	HISTORY_TRACK_PREFIX = "track"
	OBJECT_TRACK         = "object_track"
)

var db *qmgo.Database

var droneColl *qmgo.Collection
var trackHistoryColl *qmgo.Collection
var objectTrackColl *qmgo.Collection

func initColl() {
	droneColl = db.Collection(DRONE)
	objectTrackColl = db.Collection(OBJECT_TRACK)
	// trackHistoryColl = db.Collection(track_history)
	createIndex(reflect.TypeOf(pb.Drone{}), droneColl)
	createIndex(reflect.TypeOf(pb.Drone{}), objectTrackColl)

	// createIndex(reflect.TypeOf(pb.DroneTrack{}), trackHistoryColl)

}

type MainService struct {
	DbClient          *qmgo.Client
	gClient           *gclient.Client
	SvcConfig         *config.ServiceConfig
	scheduler         *gocron.Scheduler
	NATSConnection    *nats.Conn
	notifier          *Notifier
	infringedMu       sync.Mutex
	notifiedTracks    map[int32]time.Time
	activeContainment map[int32]struct{}
	tacticalMu        sync.Mutex
	notifiedConflicts map[conflictKey]time.Time
}

func createIndex(rType reflect.Type, collection *qmgo.Collection) {
	ctx := log.Logger.WithContext(context.Background())

	typeName := rType.Name()
	collectionName := collection.GetCollectionName()

	config.PrintDebugLog(ctx, "Create index for type: %s - Collection: %s", typeName, collectionName)

	for i := 0; i < rType.NumField(); i++ {
		tag := rType.Field(i).Tag

		compound := []string{
			tag.Get("bson"),
		}

		if tag.Get("compound_with") != "" {
			compound = append(compound, strings.Split(tag.Get("compound_with"), ",")...)
		}

		unique := tag.Get("index") == "unique"

		config.PrintDebugLog(ctx, "Compound: %+v - Unique: %v", compound, unique)

		if unique {
			collection.CreateOneIndex(context.Background(), options.IndexModel{
				Key:          compound,
				IndexOptions: &moptions.IndexOptions{Unique: &unique},
			})
		}
	}

	config.PrintDebugLog(ctx, "Done to create index for type: %s - Collection: %s", typeName, collectionName)
}

func (us *MainService) publishEvent(ctx context.Context, data []byte, routingKey string) {
	err := us.NATSConnection.Publish(routingKey, data)
	if err != nil {
		config.PrintErrorLog(ctx, err, "Failed to publish to nats server for: %s", routingKey)
	} else {
		config.PrintDebugLog(ctx, "Success to publish to nats server for: %s", routingKey)
	}
}

func New(dbClient *qmgo.Client, cfg config.ServiceConfig, gc *gclient.Client, nc *nats.Conn) *MainService {
	db = dbClient.Database(cfg.DbConfig.DBName)

	initColl()

	return &MainService{
		DbClient:          dbClient,
		gClient:           gc,
		SvcConfig:         &cfg,
		scheduler:         gocron.NewScheduler(time.UTC),
		NATSConnection:    nc,
		notifier:          NewNotifier(),
		notifiedTracks:    make(map[int32]time.Time),
		activeContainment: make(map[int32]struct{}),
		notifiedConflicts: make(map[conflictKey]time.Time),
	}
}

func (s *MainService) Notifier() *Notifier {
	return s.notifier
}
