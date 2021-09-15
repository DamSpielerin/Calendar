package main

import (
	"fmt"
	"log"
	"net/http"

	"calendar/server"
	"calendar/storage"

	"github.com/gookit/config/v2"
	"github.com/gookit/config/v2/yaml"
)

func main() {
	config.WithOptions(config.ParseEnv)

	// add driver for support yaml content
	config.AddDriver(yaml.Driver)

	err := config.LoadFiles("config.yaml")
	if err != nil {
		panic(err)
	}

	idle := config.Int("idle")
	pool := config.Int("port")
	port := config.Int("port")
	user := config.String("user")
	password := config.String("password")
	host := config.String("host")
	store, err := storage.NewDbStorage(fmt.Sprintf("%s:%s@tcp(%S:%d)/calendar?charset=utf8mb4&parseTime=true", user, password, host, port), idle, pool)
	if err != nil {
		log.Fatal("can't connect to db: ", err)
	}
	eventServer := server.NewEventServer(store)
	metricServer := server.NewMetricsServer(store)
	go func() {
		log.Fatal(http.ListenAndServe(":5000", eventServer))
	}()
	log.Fatal(http.ListenAndServe(":5050", metricServer))
}
