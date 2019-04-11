package main

import (
	"fetch/internal"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
)

func main() {
	tasksMaxCount := getEnvInt("MAX_TASKS_IN_QUEUE", 100)
	workersCount := getEnvInt("WORKERS_COUNT", runtime.NumCPU())

	internal.SetClient(http.DefaultClient)
	storage := internal.NewMemoryStorage()
	sc := internal.NewScheduler(tasksMaxCount, workersCount, storage)
	log.Fatalln(http.ListenAndServe(":8888", NewHandler(sc)))
}


func getEnvInt(key string, defaultValue int) int {
	val, exists := os.LookupEnv(key)

	if !exists{
		return defaultValue
	}
	v, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return v
}

