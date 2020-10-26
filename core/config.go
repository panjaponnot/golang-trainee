package core

import (
	"crypto/rsa"
	"sale_ranking/pkg/attendant"
	"sale_ranking/pkg/cache"
	"sale_ranking/pkg/crontab"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/one/chat"
	"sale_ranking/pkg/one/identity"
	"sale_ranking/pkg/util"

	"github.com/dgrijalva/jwt-go"
	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

type RSA struct {
	Key *rsa.PrivateKey
}

type Claims struct {
	*jwt.StandardClaims
}

const (
	// System Environment variable
	// Database
	envDbHost     = "DB_HOST"
	envDbPort     = "DB_PORT"
	envDbUsername = "DB_USERNAME"
	envDbPassword = "DB_PASSWORD"
	envDbName     = "DB_NAME"

	// attendant
	envAttendantToken     = "ATTENDANT_TOKEN"
	envAttendantTokenType = "ATTENDANT_TOKEN_TYPE"

	// Redis
	envRedisHost = "REDIS_HOST"
	envRedisPort = "REDIS_PORT"
	envRedisDb   = "REDIS_DB"

	// Security
	envServerKey           = "SERVER_KEY"
	envReCaptChaSecret     = "RECAPTCHA_SECRET"
	envReCaptChaTrustScore = "RECAPTCHA_TRUST_SCORE"

	// One Platform
	// Identity
	envIdClientId     = "ONEID_CLIENT_ID"
	envIdClientSecret = "ONEID_CLIENT_SECRET"

	// Chat
	envChatBotId        = "CHATBOT_ID"
	envChatBotToken     = "CHATBOT_TOKEN"
	envChatBotTokenType = "CHATBOT_TOKEN_TYPE"

	// Cache key
	// Security
	signKeyCache   = "sign_key"
	serverKeyCache = "server_key"
)

var (
	tables    []interface{}
	serverKey = util.GetEnv(envServerKey, "")
	signKey   RSA

	identityClient identity.Identity
	chatBotClient  chat.Chat

	attendantToken     = util.GetEnv(envAttendantToken, "")
	attendantTokenType = util.GetEnv(envAttendantTokenType, "Bearer")
	attendantClient    attendant.Client

	reCaptChaSiteSecret = util.GetEnv(envReCaptChaSecret, "")
	reCaptChaTrustScore = util.GetEnv(envReCaptChaTrustScore, "")
	reCaptCha           recaptcha.ReCAPTCHA
)

func NewDatabase(packageName string) database.Database {
	return NewDatabaseWithConfig(getDatabaseConfig(packageName), database.DriverMySQL)
}

func NewDatabaseWithConfig(cfg database.Config, driver string) database.Database {
	return database.New(cfg, driver)
}

func NewRedis() cache.Redis {
	return NewRedisWithConfig(getRedisConfig())
}

func NewRedisWithConfig(cfg cache.Config) cache.Redis {
	return cache.New(
		cfg.Host,
		cfg.Port,
		cfg.Db,
	)
}

func IdentityClient() *identity.Identity {
	return &identityClient
}

func ChatBotClient() *chat.Chat {
	return &chatBotClient
}

func ReCaptCha() *recaptcha.ReCAPTCHA {
	return &reCaptCha
}

func ReCaptChaTrustScore() float32 {
	return float32(util.AtoF(reCaptChaTrustScore, 0.4))
}

func AttendantClient() *attendant.Client {
	return &attendantClient
}

func CronService() *crontab.Service {
	return &cronService
}

func getDatabaseConfig(packageName string) database.Config {
	return database.Config{
		Host:        util.GetEnv(envDbHost, "127.0.0.1"),
		Port:        util.GetEnv(envDbPort, "3306"),
		Username:    util.GetEnv(envDbUsername, ""),
		Password:    util.GetEnv(envDbPassword, ""),
		Name:        util.GetEnv(envDbName, ""),
		Prod:        util.IsProduction(),
		PackageName: packageName,
	}
}

func getRedisConfig() cache.Config {
	return cache.Config{
		Host: util.GetEnv(envRedisHost, "127.0.0.1"),
		Port: util.GetEnv(envRedisPort, "6379"),
		Db:   util.AtoI(util.GetEnv(envRedisDb, "0"), 0),
	}
}

func initIdentityConfig() identity.Identity {
	return identity.NewIdentity(
		util.GetEnv(envIdClientId, ""),
		util.GetEnv(envIdClientSecret, ""),
	)
}

func initChatBotConfig() chat.Chat {
	return chat.NewChatBot(
		util.GetEnv(envChatBotId, ""),
		util.GetEnv(envChatBotToken, ""),
		util.GetEnv(envChatBotTokenType, "Bearer"),
	)
}
