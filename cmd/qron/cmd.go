package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/spf13/viper"

	"github.com/mak73kur/qron"
	"github.com/mak73kur/qron/loaders"
	"github.com/mak73kur/qron/writers"
)

func init() {
	confPath := *flag.String("c", "./qron.yml", "Path to the config file")
	flag.Parse()
	viper.SetConfigFile(confPath)
}

func requireConf(args ...string) error {
	for _, arg := range args {
		if !viper.IsSet(arg) {
			return fmt.Errorf("Config is missing required parameter: %s", arg)
		}
	}
	return nil
}

func createLoader() (qron.Loader, error) {
	if err := requireConf("loader.type"); err != nil {
		return nil, err
	}

	switch viper.GetString("loader.type") {

	case "inline":
		if err := requireConf("loader.tab"); err != nil {
			return nil, err
		}
		return loaders.Inline{viper.GetString("loader.tab")}, nil

	case "file":
		if err := requireConf("loader.path"); err != nil {
			return nil, err
		}
		return loaders.File{viper.GetString("loader.path")}, nil

	case "redis":
		if err := requireConf("loader.url", "loader.key"); err != nil {
			return nil, err
		}
		loader, err := loaders.NewRedis(viper.GetString("loader.url"))
		if err != nil {
			return nil, err
		}
		loader.Key = viper.GetString("loader.key")

		if viper.IsSet("loader.auth") {
			if err = loader.Auth(viper.GetString("loader.auth")); err != nil {
				return nil, err
			}
		}
		if viper.IsSet("loader.db") {
			if err := loader.Select(viper.GetInt("loader.db")); err != nil {
				return nil, err
			}
		}
		return loader, nil

	default:
		return nil, fmt.Errorf("unknown loader type: %s", viper.GetString("loader.type"))
	}
}

func createWriter() (qron.Writer, error) {
	if err := requireConf("writer.type"); err != nil {
		return nil, err
	}

	switch viper.GetString("writer.type") {

	case "log":
		return writers.Log{}, nil

	case "amqp":
		if err := requireConf("writer.url", "writer.exchange", "writer.routing_key"); err != nil {
			return nil, err
		}
		return writers.NewAMQP(
			viper.GetString("writer.url"),
			viper.GetString("writer.exchange"),
			viper.GetString("writer.routing_key"))

	case "redis":
		if err := requireConf("writer.url", "writer.key"); err != nil {
			return nil, err
		}
		writer, err := writers.NewRedis(viper.GetString("writer.url"))
		if err != nil {
			return nil, err
		}
		writer.Key = viper.GetString("writer.key")
		writer.LeftPush = viper.GetBool("writer.left_push")

		if viper.IsSet("writer.auth") {
			if err = writer.Auth(viper.GetString("writer.auth")); err != nil {
				return nil, err
			}
		}
		if viper.IsSet("loader.db") {
			if err := writer.Select(viper.GetInt("writer.db")); err != nil {
				return nil, err
			}
		}
		return writer, nil

	default:
		return nil, fmt.Errorf("unknown writer type: %s", viper.GetString("writer.type"))
	}
}

// stop program in case of error
func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	err := viper.ReadInConfig()
	check(err)

	loader, err := createLoader()
	check(err)

	writer, err := createWriter()
	check(err)

	// create and load schedule
	sch := qron.NewSchedule(loader, writer)
	check(sch.LoadAndWatch())

	// start schedule ticker
	sch.Run()
}
