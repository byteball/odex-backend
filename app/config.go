package app

import (
	"fmt"

	"github.com/byteball/odex-backend/utils"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
)

// Config stores the application-wide configurations
var Config appConfig
var logger = utils.Logger

type appConfig struct {
	// the path to the error message file. Defaults to "config/errors.yaml"
	ErrorFile string `mapstructure:"error_file"`
	// the server port. Defaults to 8080
	ServerPort int `mapstructure:"server_port"`
	// the data source name (MongoURL) for connecting to the database. required.
	MongoURL         string `mapstructure:"mongo_url"`
	MongoDBPassword  string `mapstructure:"mongo_password"`
	MongoDBUsername  string `mapstructure:"mongo_username"`
	MongoDBShardURL1 string `mapstructure:"mongo_shard_url_1"`
	MongoDBShardURL2 string `mapstructure:"mongo_shard_url_2"`
	MongoDBShardURL3 string `mapstructure:"mongo_shard_url_3"`

	RabbitMQURL      string `mapstructure:"rabbitmq_url"`
	RabbitMQPassword string `mapstructure:"rabbitmq_password"`
	RabbitMQUsername string `mapstructure:"rabbitmq_username"`

	// the data source name (MongoURL) for connecting to the database. required.
	DBName string `mapstructure:"db_name"`
	// TickDuration is user by tick streaming cron
	TickDuration map[string][]int64 `mapstructure:"tick_duration"`

	Logs map[string]string `mapstructure:"logs"`

	Obyte map[string]string `mapstructure:"obyte"`

	Env string `mapstructure:"env"`

	EnableTLS    bool   `mapstructure:"enable_tls"`
	ServerCACert string `mapstructure:"server_ca_cert"`
	ServerCert   string `mapstructure:"server_cert"`
	ServerKey    string `mapstructure:"server_key"`
	MongoDBKey   string `mapstructure:"mongodb_key"`
	MongoDBCert  string `mapstructure:"mongodb_cert"`
	RabbitMQKey  string `mapstructure:"rabbitmq_cert"`
	RabbitMQCert string `mapstructure:"rabbitmq_key"`
}

func (config appConfig) Validate() error {
	return validation.ValidateStruct(&config,
		validation.Field(&config.MongoURL, validation.Required),
	)
}

// LoadConfig loads configuration from the given list of paths and populates it into the Config variable.
// The configuration file(s) should be named as app.yaml.
// Environment variables with the prefix "RESTFUL_" in their names are also read automatically.
func LoadConfig(configPath string, env string) error {
	v := viper.New()

	if env != "" {
		v.SetConfigName("config." + env)
	}

	v.SetConfigType("yaml")
	v.AddConfigPath(configPath)

	err := v.ReadInConfig()
	if err != nil {
		return fmt.Errorf("Failed to read the configuration file: %s", err)
	}

	err = v.Unmarshal(&Config)
	if err != nil {
		return err
	}

	v.SetEnvPrefix("odex")
	v.AutomaticEnv()

	//General Configuration
	Config.Env = env
	Config.ServerPort = v.Get("SERVER_PORT").(int)
	Config.ErrorFile = "config/errors.yaml"

	//RabbitMQ Configuration
	Config.RabbitMQURL = v.Get("RABBITMQ_URL").(string)

	//Mongo Configuration
	Config.MongoURL = v.Get("MONGODB_URL").(string)
	Config.DBName = v.Get("MONGODB_DBNAME").(string)

	//TLS/SSL Configuration
	tlsEnabled := v.Get("ENABLE_TLS").(string)
	if tlsEnabled == "true" {
		Config.EnableTLS = true
		Config.ServerCACert = v.Get("MATCHING_ENGINE_CA_CERT").(string)
		Config.ServerCert = v.Get("MATCHING_ENGINE_SERVER_CERT").(string)
		Config.ServerKey = v.Get("MATCHING_ENGINE_SERVER_KEY").(string)
		Config.RabbitMQKey = v.Get("RABBITMQ_CLIENT_KEY").(string)
		Config.RabbitMQCert = v.Get("RABBITMQ_CLIENT_CERT").(string)
		Config.MongoDBUsername = v.Get("MONGODB_USERNAME").(string)
		Config.MongoDBPassword = v.Get("MONGODB_PASSWORD").(string)
		Config.RabbitMQUsername = v.Get("RABBITMQ_USERNAME").(string)
		Config.MongoDBShardURL1 = v.Get("MONGODB_SHARD_URL_1").(string)
		Config.RabbitMQPassword = v.Get("RABBITMQ_PASSWORD").(string)
		Config.MongoDBShardURL2 = v.Get("MONGODB_SHARD_URL_2").(string)
		Config.MongoDBShardURL3 = v.Get("MONGODB_SHARD_URL_3").(string)
	} else {
		Config.EnableTLS = false
		Config.ServerCACert = ""
		Config.ServerCert = ""
		Config.ServerKey = ""
		Config.RabbitMQKey = ""
		Config.RabbitMQCert = ""
		Config.MongoDBUsername = ""
		Config.MongoDBPassword = ""
		Config.RabbitMQUsername = ""
		Config.RabbitMQPassword = ""
		Config.MongoDBShardURL1 = ""
		Config.MongoDBShardURL2 = ""
		Config.MongoDBShardURL3 = ""
	}

	Config.Obyte = make(map[string]string)
	Config.Obyte["http_url"] = v.Get("OBYTE_NODE_HTTP_URL").(string)
	Config.Obyte["ws_url"] = v.Get("OBYTE_NODE_WS_URL").(string)

	logger.Infof("Server port: %v", Config.ServerPort)
	logger.Infof("Obyte node HTTP url: %v", Config.Obyte["http_url"])
	logger.Infof("Obyte node WS url: %v", Config.Obyte["ws_url"])
	logger.Infof("MongoDB url: %v", Config.MongoURL)
	logger.Infof("MongoDB db name: %v", Config.DBName)
	logger.Infof("MongoUserName: %v", Config.MongoDBUsername)
	logger.Infof("MongoShardURL2: %v", Config.MongoDBShardURL1)
	logger.Infof("MongoShardURL2: %v", Config.MongoDBShardURL2)
	logger.Infof("MongoShardURL2: %v", Config.MongoDBShardURL3)
	logger.Infof("RabbitMQ url: %v", Config.RabbitMQURL)
	logger.Infof("RabbitMQUserName: %v", Config.RabbitMQUsername)
	logger.Infof("TLS Enabled: %v", Config.EnableTLS)

	return Config.Validate()
}
