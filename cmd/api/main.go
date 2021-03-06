package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"ozon_test/app/handlers"
	"ozon_test/app/middleware"
	"ozon_test/app/repositories"
	"ozon_test/app/repositories/impl"
	"ozon_test/app/usecases/impl"
	"strings"
)

type Flags struct {
	PostgresDatabase bool
}

func ParseFlag() (flags Flags) {
	flag.BoolVar(&flags.PostgresDatabase, "postgres", false, "use postgresql database")
	flag.Parse()
	return
}

type Config struct {
	postgresHost     string
	postgresUser     string
	postgresPassword string
	postgresDbName   string
	postgresPort     string

	logFile string
}

func ParseConfig() (conf Config) {
	viper.AddConfigPath("./cmd/api/")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	conf.postgresHost = viper.GetString("postgresHost")
	conf.postgresUser = viper.GetString("postgresUser")
	conf.postgresPassword = viper.GetString("postgresPassword")
	conf.postgresDbName = viper.GetString("postgresDbName")
	conf.postgresPort = viper.GetString("postgresPort")

	conf.logFile = viper.GetString("logFile")
	return
}

func main() {
	flags := ParseFlag()

	conf := ParseConfig()
	f, err := os.Create(conf.logFile)
	if err != nil {
		log.Fatal(err)
		return
	}

	gin.DefaultWriter = io.MultiWriter(f)
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST"}
	config.AllowCredentials = true

	var linkRepository repositories.LinkRepository

	if flags.PostgresDatabase {
		conn, err := pgx.ParseConnectionString(strings.Join([]string{"host=", conf.postgresHost, " user=", conf.postgresUser, " password=", conf.postgresPassword, " dbname=", conf.postgresDbName, " port=", conf.postgresPort}, ""))
		if err != nil {
			log.Fatal(err)
			return
		}

		db, err := pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     conn,
			MaxConnections: 100,
			AfterConnect:   nil,
			AcquireTimeout: 0,
		})
		if err != nil {
			log.Fatal(err)
			return
		}

		defer db.Close()
		linkRepository = repositories_impl.MakePostgresRepository(db)
	} else {
		linkRepository = repositories_impl.MakeInMemoryRepository()
	}

	linkUseCase := usecases_impl.MakeLinkUseCase(linkRepository)
	linkHandler := handlers.MakeLinkHandler(linkUseCase)

	router.Use(cors.New(config))
	router.Use(middleware.CheckError())

	routes := router.Group("/")
	{
		routes.GET("/:link", linkHandler.GetLink)
		routes.POST("/:link", linkHandler.CreateLink)
	}

	err = router.Run(":5000")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
}
