package config

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type ServiceConfig struct {
	DbConfig          MongoConfig             `mapstructure:"mongo"`
	GrpcConfig        GrpcConfig              `mapstructure:"grpc"`
	HttpConfig        HttpConfig              `mapstructure:"http"`
	LoggerConfig      LoggerConfig            `mapstructure:"logger"`
	RabbitmqConfig    RabbitMQConfig          `mapstructure:"rabbitmq"`
	OtherConfig       OtherConfig             `mapstructure:"other"`
	NATSConfig        NATSConfig              `mapstructure:"nats"`
	JWTTokenConfig    JWTTokenConfig          `mapstructure:"jwt_token_config"`
	FlightContainment FlightContainmentConfig `mapstructure:"flight_containment"`
}

func LoadConfig(path string) (cfg ServiceConfig, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.AutomaticEnv()

	setDefaultValue()

	err = viper.ReadInConfig()
	if err != nil {
		log.Fatal().Msgf("Failed to load config: %v", err)

		return
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal().Msgf("Failed to mapping config: %v", err)

		return
	}

	return
}

func setDefaultValue() {
	/* Config mongo db */
	viper.SetDefault("mongo.host", "localhost")
	viper.SetDefault("mongo.port", 27017)
	viper.SetDefault("mongo.name", "mongo")
	viper.SetDefault("mongo.user", "mongo")
	viper.SetDefault("mongo.pass", "mongo")

	/* Config http */
	viper.SetDefault("http.port", 8080)
	viper.SetDefault("http.enable_recover_middleware", true)
	viper.SetDefault("http.enable_cors_middleware", true)

	httpEndpoints := map[string]string{
		SVC_EVENT_LISTENER: "localhost:33980",
		SVC_ORDER:          "localhost:33180",
		SVC_DRONE:          "localhost:33280",
		SVC_COMMAND:        "localhost:33380",
		SVC_LOCKER:         "localhost:33480",
		SVC_USER:           "localhost:33580",
		SVC_GCS:            "localhost:33680",
		SVC_DATASOURCE:     "localhost:33780",
	}
	viper.SetDefault("http.endpoints", httpEndpoints)

	/* Config grpc */
	viper.SetDefault("grpc.port", 8065)

	grpcChannels := map[string]string{
		SVC_EVENT_LISTENER: "localhost:33965",
		SVC_ORDER:          "localhost:33165",
		SVC_DRONE:          "localhost:33265",
		SVC_COMMAND:        "localhost:33365",
		SVC_LOCKER:         "localhost:33465",
		SVC_USER:           "localhost:33565",
		SVC_GCS:            "localhost:33665",
		SVC_DATASOURCE:     "localhost:33765",
	}
	viper.SetDefault("grpc.channels", grpcChannels)

	/* Config rabbitmq */
	rabbitmqMatchingPatterns := map[string][]string{
		"system-event": {"*.*.*.*.#"},
	}

	viper.SetDefault("rabbitmq.username", "admin")
	viper.SetDefault("rabbitmq.password", "admin")
	viper.SetDefault("rabbitmq.vhost", "/")
	viper.SetDefault("rabbitmq.schema", "amqp")
	viper.SetDefault("rabbitmq.reconnect_max_attempt", "100")
	viper.SetDefault("rabbitmq.reconnect_interval", "5")
	viper.SetDefault("rabbitmq.channel_timeout", "5")
	viper.SetDefault("rabbitmq.event_exchange", "system-event")
	viper.SetDefault("rabbitmq.event_listen_routing_key", rabbitmqMatchingPatterns)

	/* Config nats */
	natsMatchingPattern := []string{"*.*.*.*.*"}
	viper.SetDefault("nats.listen_to_subject", natsMatchingPattern)

	/* Config jwt */
	viper.SetDefault("jwt_token_config.validate_jwt", false)

	/* Config other */
	viper.SetDefault("other.environment", "development")
	viper.SetDefault("other.default_lang", "en")
	viper.SetDefault("other.bundle_dir_abs", "web/i18n")
	viper.SetDefault("other.tracing_host", "localhost")
	viper.SetDefault("other.tracing_port", 9411)

	viper.SetDefault("flight_containment.radius", 5.0)
	viper.SetDefault("flight_containment.waypoints", []map[string]float64{
		{"lat": 21.002694, "lon": 105.537611, "alt": 40},
		{"lat": 21.001444, "lon": 105.538111, "alt": 40},
		{"lat": 21.0005, "lon": 105.535889, "alt": 40},
		{"lat": 21.001944, "lon": 105.535222, "alt": 40},
	})
	viper.SetDefault("flight_containment.renotify_seconds", 60.0)
	viper.SetDefault("flight_containment.horizontal_deviation_m", 10.0)
	viper.SetDefault("flight_containment.alt_deviation_m", 5.0)
}
