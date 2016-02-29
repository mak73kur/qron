package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"

	"github.com/mak73kur/qron"
	"github.com/mak73kur/qron/loaders"
	"github.com/mak73kur/qron/writers"
)

// stop program in case of error
func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func require(args ...string) {
	for _, arg := range args {
		if !viper.IsSet(arg) {
			check(fmt.Errorf("Config is missing required parameter: %s", arg))
		}
	}
}

func init() {
	if len(os.Args) > 1 {
		viper.SetConfigFile(os.Args[1])
	} else {
		viper.SetConfigFile("/etc/qron.yml")
	}
	err := viper.ReadInConfig()
	check(err)
}

func main() {
	require("loader.type", "writer.type")

	var (
		loader qron.Loader
		writer qron.Writer
		err    error
	)

	switch viper.GetString("loader.type") {
	case "inline":
		require("loader.tab")
		loader = loaders.NewInline(viper.GetString("loader.tab"))
	case "file":
		require("loader.path")
		loader = loaders.NewFile(viper.GetString("loader.path"))
	case "redis":
		require("loader.url", "loader.key")

		loader, err = loaders.NewRedis(viper.GetString("loader.url"), viper.GetString("loader.key"))
		check(err)

		if viper.IsSet("loader.db") {
			err = loader.(*loaders.Redis).Select(viper.GetInt("db"))
			check(err)
		}
		if viper.IsSet("loader.password") {
			err = loader.(*loaders.Redis).Auth(viper.GetString("loader.password"))
			check(err)
		}
	default:
		check(fmt.Errorf("unknown loader type: %s", viper.GetString("loader.type")))
	}

	switch viper.GetString("writer.type") {
	case "log":
		writer = writers.Log{}
	case "amqp":
		require("writer.url", "writer.exchange", "writer.routing_key")
		writer, err = writers.NewAMQP(viper.GetString("writer.url"), viper.GetString("writer.exchange"),
			viper.GetString("writer.routing_key"))
		check(err)
	default:
		check(fmt.Errorf("unknown writer type: %s", viper.GetString("writer.type")))
	}

	// create and load schedule
	sch := qron.NewSchedule(loader, writer)
	check(sch.LoadAndWatch())

	// start schedule ticker
	sch.Run()
}
