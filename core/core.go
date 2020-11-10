package core

import (
	// " golang_template/pkg/attendant"

	m "sale_ranking/model"
	"sale_ranking/pkg/cache"
	"sale_ranking/pkg/crontab"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/log"
	"time"

	"gopkg.in/ezzarghili/recaptcha-go.v4"
)

const pkgName = "CORE"

var (
	dbSale     database.Database
	dbQuataion database.Database
	dbMssql    database.Database
	redis      cache.Redis

	cronService crontab.Service
)

func init() {
	// Preparing database schema
	tables = []interface{}{
		// client
		&m.ApiClient{},
	}
}

func InitCoreService() error {
	// Database Sale
	dbSale = NewDatabase(pkgName, "salerank")
	if err := dbSale.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database error")
		return err
	}
	log.Infoln(pkgName, "Connected to database sale ranking.")
	// Database Quataion
	dbQuataion = NewDatabase(pkgName, "quotation")
	if err := dbQuataion.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database error")
		return err
	}
	log.Infoln(pkgName, "Connected to database quotation.")
	dbMssql = NewDatabaseMssql(pkgName, "mssql")
	if err := dbMssql.Connect(); err != nil {
		log.Errorln(pkgName, err, "Connect to database sql server error")
		return err
	}
	log.Infoln(pkgName, "Connected to database sql server.")
	// Migrate database
	// if err := db.MigrateDatabase(tables); err != nil {
	// 	log.Errorln(pkgName, err, "Migrate database error")
	// 	return err
	// }
	// log.Infoln(pkgName, "Migrated database schema.")

	// Redis cache
	redis = NewRedis()
	if err := redis.Ping(); err != nil {
		log.Errorln(pkgName, err, "Connect to redis error")
		return err
	}
	log.Infoln(pkgName, "Connected to redis server.")

	// Init security key
	if err := InitSecurityKey(); err != nil {
		log.Errorln(pkgName, err, "Initializing system security error")
		return err
	}
	log.Infoln(pkgName, "Initialized system security key.")

	// Prepare reCapCha
	var err error
	reCaptCha, err = recaptcha.NewReCAPTCHA(reCaptChaSiteSecret, recaptcha.V3, 10*time.Second)
	if err != nil {
		log.Errorln(pkgName, err, "New ReCaptCha error")
		return err
	}
	log.Infoln(pkgName, "New reCaptCha success.")

	// Prepare AttendantClient
	// attendantClient, err = attendant.NewClient(attendantToken, attendantTokenType)
	// if err != nil {
	// 	log.Errorln(pkgName, err, "New Attendant client error")
	// 	return err
	// }
	// log.Infoln(pkgName, "New Attendant client success.")

	// Prepare one platform config
	// Identity
	identityClient = initIdentityConfig()
	// Chat
	chatBotClient = initChatBotConfig()
	log.Infoln(pkgName, "Initialized one platform client.")

	// Init crontab service
	// cronService = crontab.NewService()
	// cronService.Start()
	// log.Infoln(pkgName, "Crontab service started.")

	return nil
}
