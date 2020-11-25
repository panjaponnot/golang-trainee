package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"sale_ranking/api"
	"sale_ranking/auth"
	"sale_ranking/core"
	"sale_ranking/pkg/database"
	"sale_ranking/pkg/log"
	"sale_ranking/pkg/server"
	"sale_ranking/pkg/util"

	"github.com/labstack/echo/v4"
	"gopkg.in/alecthomas/kingpin.v2"
)

const (
	pkgName     = "MAIN"
	ProgramName = "Sale Ranking"
	Version     = "1.0.0"
)

var clientArgs = struct {
	createName *string
	delName    *string
}{}

func main() {
	// Variable Parameter
	var (
		startArgs = struct {
			host   *net.IP
			port   *string
			prefix *string
			prod   *bool
		}{}
	)

	// Command line
	a := kingpin.New(filepath.Base(os.Args[0]), fmt.Sprintf("%s %s", ProgramName, Version))
	a.Version(Version)
	a.HelpFlag.Short('h')

	// Start
	startCmd := a.Command("start", "Start server command.")
	startArgs.host = startCmd.Flag("host", "Set server host address.").Envar("SERVER_HOST").Default("0.0.0.0").IP()
	startArgs.port = startCmd.Flag("port", "Set server listen port.").Envar("SERVER_PORT").Default("5000").String()
	startArgs.prefix = startCmd.Flag("prefix", "Server api prefix.").Envar("SERVER_PREFIX").Default("/api").String()

	// token client
	clientCmd := a.Command("client", "Api client management.")
	lsClientCmd := clientCmd.Command("ls", "List available client.")
	createClientCmd := clientCmd.Command("create", "Create new client.")
	clientArgs.createName = createClientCmd.Arg("name", "Client name.").Required().String()
	delClientCmd := clientCmd.Command("del", "Delete api client.")
	clientArgs.delName = delClientCmd.Arg("name", "Client name.").Required().String()

	// sync billing
	billingCmd := a.Command("sync", "Sync billing management.")

	startTime := time.Now()
	// Init core service
	if err := core.InitCoreService(); err != nil {
		_ = cleanUp()
		util.ExitWithCode(startTime, 100)
	}
	switch kingpin.MustParse(a.Parse(os.Args[1:])) {
	case startCmd.FullCommand():
		// $ start --host=HOST --port=PORT --prefix=PREFIX
		log.Infoln(pkgName, "===== Template Server Golang v.", Version, " starting at", time.Now().Format(time.ANSIC), " Production:", util.IsProduction(), "=====")
		s := server.New(server.Config{
			Host:   startArgs.host.String(),
			Port:   *startArgs.port,
			Prefix: *startArgs.prefix,
			Prod:   util.IsProduction(),
		})
		if err := initApiRouter(s.Ctx()); err != nil {
			log.Errorln(pkgName, err, "Init API router error")
			_ = cleanUp()
			util.ExitWithCode(startTime, 101)
		}
		if err := s.Run(); err != nil {
			log.Errorln(pkgName, err, "Start server error")
			_ = cleanUp()
			util.ExitWithCode(startTime, 102)
		}
		_ = cleanUp()
		log.Infoln(pkgName, "Server terminated")
	case lsClientCmd.FullCommand():
		// - client ls
		exitWithCode(startTime, core.GetApiClientListCli())
	case createClientCmd.FullCommand():
		// - client create <name>
		exitWithCode(startTime, core.AddApiClient(*clientArgs.createName))
	case delClientCmd.FullCommand():
		// - client del <name>
		exitWithCode(startTime, core.DeleteApiClient(*clientArgs.delName))
	case billingCmd.FullCommand():
		log.Infoln(pkgName, "===========  Synchronize Billing ===========")
		exitWithCode(startTime, core.SyncBillingToDB())
	}
}

func exitWithCode(startTime time.Time, code int) {
	log.Infoln("Elapsed time", time.Since(startTime).Seconds(), "second(s).")
	os.Exit(code)
}

func cleanUp() error {
	// Database
	_ = cleanUpDatabase()
	// Stop crontab service
	// clearUpCronService()
	return nil
}

func cleanUpDatabase() error {
	log.Infoln(pkgName, "Cleaning up database connection session.")
	cleanSuccess := 0
	for _, db := range database.GetConnectionContext() {
		if err := db.Close(); err != nil {
			log.Errorln(pkgName, err, "Close database session error")
		} else {
			cleanSuccess += 1
		}
	}
	log.Infoln(pkgName, "Clean up database connection success", cleanSuccess, "/", len(database.GetConnectionContext()))
	return nil
}

func clearUpCronService() {
	log.Infoln(pkgName, "Clean up register cron service.")
	core.CronService().Stop()
}

func initApiRouter(ctx *echo.Echo) error {
	apiV2 := ctx.Group("/api/v2")
	// Authentication - api
	if err := auth.InitApiRouter(apiV2.Group("/auth")); err != nil {
		return err
	}

	// User Management  - api
	if err := api.InitApiRouter(apiV2.Group("")); err != nil {
		return err
	}

	// Transaction - api
	// if err := transaction.InitAPIRouter(apiV1.Group("/transaction")); err != nil {
	// 	return err
	// }

	return nil
}
