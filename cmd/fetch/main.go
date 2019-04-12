package main

import (
	"context"
	"fetch/internal"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

)

// @title Tasks API
// @version 1.0
// @description Документация для Tasks API

// @contact.name API Support
// @contact.email matrosov.nikita94@gmail.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8888
// @BasePath /

func main() {
	addr := getEnv("SERVER_ADDRESS", ":8888")
	tasksMaxCount := getEnvInt("MAX_TASKS_IN_QUEUE", 100)
	workersCount := getEnvInt("WORKERS_COUNT", runtime.NumCPU())

	internal.SetClient(http.DefaultClient)
	storage := internal.NewMemoryStorage()
	sc := internal.NewScheduler(tasksMaxCount, workersCount, storage)

	r := mux.NewRouter()
	h := NewHandler(sc)
	h.Attach(r)

	srv := &http.Server{
		Addr:         addr,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: r,
	}

	go func() {
		log.Println("Listening on", addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Println(err)
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt)

	<-shutdownCh

	ctx, cancel := context.WithTimeout(context.Background(), 15 *time.Second)
	defer cancel()

	srv.Shutdown(ctx)
	os.Exit(0)
}

func getEnv(name string, defaultVal string) string {
	if value, exists := os.LookupEnv(name); exists {
		return value
	}

	return defaultVal
}

func getEnvInt(name string, defaultValue int) int {
	valueStr := getEnv(name, "")
	v, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return v
}

