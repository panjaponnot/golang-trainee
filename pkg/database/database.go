package database

import (
	"fmt"
	"sale_ranking/pkg/log"

	"github.com/carlescere/scheduler"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Config struct {
	Host        string
	Port        string
	Username    string
	Password    string
	Name        string
	Prod        bool
	PackageName string
	Filename    string
}

type Database struct {
	config Config
	driver string
	dsn    string
	ctx    *gorm.DB
	job    *scheduler.Job
}

const (
	DriverMSSQL    = "mssql"
	DriverMySQL    = "mysql"
	DriverSQLLite  = "sqlite3"
	DriverPostgres = "postgres"
)

var dbContext []*gorm.DB

func GetConnectionContext() []*gorm.DB {
	return dbContext
}

func New(cfg Config, driver string) Database {
	return Database{
		config: cfg,
		driver: driver,
	}
}

func (db *Database) Connect() error {
	var err error
	var dsn string
	var driver string
	switch db.driver {
	case DriverMSSQL:
		driver = DriverMSSQL
		dsn = fmt.Sprintf(
			// "sqlserver://%s:%s@%s:%s?database=%s",
			"server=%s;user id=%s;password=%s;database=%s",
			db.config.Host,
			db.config.Username,
			db.config.Password,
			// db.config.Port,
			db.config.Name,
		)
	case DriverPostgres:
		driver = DriverPostgres
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s dbname=%s password=%s",
			db.config.Host,
			db.config.Port,
			db.config.Username,
			db.config.Name,
			db.config.Password,
		)
	case DriverSQLLite:
		driver = DriverSQLLite
		dsn = db.config.Filename
	default:
		driver = DriverMySQL
		// dsn = fmt.Sprintf(
		// 	"root:mis@Pass01@tcp(203.151.56.242:3306)/ratingscoring?charset=utf8mb4&parseTime=true&loc=Local",
		// )
		dsn = fmt.Sprintf(
			`%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true&loc=Local`,
			db.config.Username,
			db.config.Password,
			db.config.Host,
			db.config.Port,
			db.config.Name,
		)
	}
	// log.Infoln(db.config.Username, " +=======+ ")
	// log.Infoln(dsn, " +=======+ ", db.driver)
	db.dsn = dsn
	db.ctx, err = gorm.Open(driver, db.dsn)
	if err != nil {
		return err
	}
	if err := db.startKeepAlive(); err != nil {
		return err
	}
	db.ctx.LogMode(!db.config.Prod)
	dbContext = append(dbContext, db.ctx)
	return nil
}

func (db *Database) Reconnect() error {
	var err error
	db.ctx, err = gorm.Open(db.driver, db.dsn)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) Ctx() *gorm.DB {
	return db.ctx
}

func (db *Database) MigrateDatabase(tables []interface{}) error {
	tx := db.ctx.Begin()
	for _, t := range tables {
		if err := tx.AutoMigrate(t).Error; err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func (db *Database) startKeepAlive() error {
	var err error
	db.job, err = scheduler.Every(30).Seconds().Run(func() {
		if err := db.ctx.DB().Ping(); err != nil {
			log.Errorln(db.config.PackageName, err, "Database keepalive error")
			if err := db.Reconnect(); err != nil {
				log.Errorln(db.config.PackageName, err, "Trying to reconnect to database error")
			} else {
				log.Infoln(db.config.PackageName, "Database reconnect success.")
			}
		}
	})
	return err
}

func (db *Database) stopKeepAlive() error {
	if db.job != nil {
		db.job.Quit <- true
	}
	return nil
}

func (db *Database) Close() error {
	_ = db.stopKeepAlive()
	if err := db.ctx.Close(); err != nil {
		return err
	}
	return nil
}
