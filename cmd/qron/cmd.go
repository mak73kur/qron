package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/mak73kur/qron"
	"github.com/spf13/viper"
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

func createReader() (qron.Reader, error) {
	if err := requireConf("reader.type"); err != nil {
		return nil, err
	}

	switch viper.GetString("reader.type") {

	case "inline":
		if err := requireConf("reader.tab"); err != nil {
			return nil, err
		}
		return qron.InlineReader{[]byte(viper.GetString("reader.tab"))}, nil

	case "file":
		if err := requireConf("reader.path"); err != nil {
			return nil, err
		}
		return qron.FileReader{viper.GetString("reader.path")}, nil

	case "redis":
		if err := requireConf("reader.url", "reader.key"); err != nil {
			return nil, err
		}
		reader, err := qron.NewRedisReader(
			viper.GetString("reader.url"),
			viper.GetString("reader.auth"),
			viper.GetInt("reader.db"))
		if err != nil {
			return nil, err
		}
		reader.Key = viper.GetString("reader.key")
		return reader, nil

	default:
		return nil, fmt.Errorf("unknown reader type: %s", viper.GetString("reader.type"))
	}
}

func createWriter() (qron.Writer, error) {
	if err := requireConf("writer.type"); err != nil {
		return nil, err
	}

	switch viper.GetString("writer.type") {

	case "log":
		return qron.LogWriter{}, nil

	case "amqp":
		if err := requireConf("writer.url", "writer.exchange", "writer.routing_key"); err != nil {
			return nil, err
		}
		return qron.NewAMQP(
			viper.GetString("writer.url"),
			viper.GetString("writer.exchange"),
			viper.GetString("writer.routing_key"))

	case "redis":
		if err := requireConf("writer.url", "writer.key"); err != nil {
			return nil, err
		}
		writer, err := qron.NewRedisWriter(
			viper.GetString("reader.url"),
			viper.GetString("reader.auth"),
			viper.GetInt("reader.db"))
		if err != nil {
			return nil, err
		}
		writer.Key = viper.GetString("writer.key")
		writer.LeftPush = viper.GetBool("writer.left_push")
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

	reader, err := createReader()
	check(err)

	writer, err := createWriter()
	check(err)

	// create and load schedule
	sch := qron.NewSchedule(reader, writer)
	check(sch.LoadAndWatch())

	// start schedule ticker
	sch.Run()
}
