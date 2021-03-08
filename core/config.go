package core

import (
	"crypto/rsa"
	"sale_ranking/pkg/attendant"
	"sale_ranking/pkg/billing"
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
	// envDbHost     = "DB_HOST"
	// envDbPort     = "DB_PORT"
	// envDbUsername = "DB_USERNAME"
	// envDbPassword = "DB_PASSWORD"
	// envDbName     = "DB_NAME"

	// db mssql equip
	envDbMssqlEquipHost     = "DB_MSSQL_EQUIP_HOST"
	envDbMssqlEquipPort     = "DB_MSSQL_EQUIP_PORT"
	envDbMssqlEquipUsername = "DB_MSSQL_EQUIP_USERNAME"
	envDbMssqlEquipPassword = "DB_MSSQL_EQUIP_PASSWORD"
	envDbMssqlEquipName     = "DB_MSSQL_EQUIP_NAME"

	// db sale
	envDbSaleHost     = "DB_SALE_HOST"
	envDbSalePort     = "DB_SALE_PORT"
	envDbSaleUsername = "DB_SALE_USERNAME"
	envDbSalePassword = "DB_SALE_PASSWORD"
	envDbSaleName     = "DB_SALE_NAME"
	// db quotation
	envDbQuotationHost     = "DB_QUOTATION_HOST"
	envDbQuotationPort     = "DB_QUOTATION_PORT"
	envDbQuotationUsername = "DB_QUOTATION_USERNAME"
	envDbQuotationPassword = "DB_QUOTATION_PASSWORD"
	envDbQuotationName     = "DB_QUOTATION_NAME"

	// db mssql
	envDbMssqlHost     = "DB_MSSQL_HOST"
	envDbMssqlPort     = "DB_MSSQL_PORT"
	envDbMssqlUsername = "DB_MSSQL_USERNAME"
	envDbMssqlPassword = "DB_MSSQL_PASSWORD"
	envDbMssqlName     = "DB_MSSQL_NAME"

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
	tables     []interface{}
	tablesSale []interface{}
	serverKey  = util.GetEnv(envServerKey, "")
	signKey    RSA

	identityClient identity.Identity
	chatBotClient  chat.Chat

	attendantToken     = util.GetEnv(envAttendantToken, "")
	attendantTokenType = util.GetEnv(envAttendantTokenType, "Bearer")
	attendantClient    attendant.Client

	reCaptChaSiteSecret = util.GetEnv(envReCaptChaSecret, "")
	reCaptChaTrustScore = util.GetEnv(envReCaptChaTrustScore, "")
	reCaptCha           recaptcha.ReCAPTCHA
)

func NewDatabase(packageName string, name string) database.Database {
	return NewDatabaseWithConfig(getDatabaseConfig(packageName, name), database.DriverMySQL)
}
func NewDatabaseMssql(packageName string, name string) database.Database {
	return NewDatabaseWithConfig(getDatabaseConfig(packageName, name), database.DriverMSSQL)
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

func getDatabaseConfig(packageName string, dbName string) database.Config {
	// log.Infoln("pkg", "==", util.GetEnv(envDbQuotationHost, "127.0.0.1"), "+=======---- ", dbName)
	switch dbName {
	case "quotation":
		return database.Config{
			Host:        util.GetEnv(envDbQuotationHost, "127.0.0.1"),
			Port:        util.GetEnv(envDbQuotationPort, "3306"),
			Username:    util.GetEnv(envDbQuotationUsername, ""),
			Password:    util.GetEnv(envDbQuotationPassword, ""),
			Name:        util.GetEnv(envDbQuotationName, ""),
			Prod:        util.IsProduction(),
			PackageName: packageName,
		}
	case "salerank":
		// log.Infoln("pkg", "==", util.GetEnv(envDbSaleName, ""), "+=======---- ", util.IsProduction())
		return database.Config{
			Host:        util.GetEnv(envDbSaleHost, "127.0.0.1"),
			Port:        util.GetEnv(envDbSalePort, "3306"),
			Username:    util.GetEnv(envDbSaleUsername, ""),
			Password:    util.GetEnv(envDbSalePassword, ""),
			Name:        util.GetEnv(envDbSaleName, ""),
			Prod:        util.IsProduction(),
			PackageName: packageName,
		}
	case "mssql":
		return database.Config{
			Host:        util.GetEnv(envDbMssqlHost, "127.0.0.1"),
			Port:        util.GetEnv(envDbMssqlPort, "1433"),
			Username:    util.GetEnv(envDbMssqlUsername, ""),
			Password:    util.GetEnv(envDbMssqlPassword, ""),
			Name:        util.GetEnv(envDbMssqlName, ""),
			Prod:        util.IsProduction(),
			PackageName: packageName,
		}
	case "equip":
		return database.Config{
			Host:        util.GetEnv(envDbMssqlEquipHost, "127.0.0.1"),
			Port:        util.GetEnv(envDbMssqlEquipPort, "1433"),
			Username:    util.GetEnv(envDbMssqlEquipUsername, ""),
			Password:    util.GetEnv(envDbMssqlEquipPassword, ""),
			Name:        util.GetEnv(envDbMssqlEquipName, ""),
			Prod:        util.IsProduction(),
			PackageName: packageName,
		}
	}
	return database.Config{
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

func initBillingConfig(token string) billing.Billing {
	return billing.NewBilling(
		token,
	)
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
