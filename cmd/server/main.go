package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
)

const httpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")
		fmt.Fprintf(w, "Hello, %s!", city)
		_, err := w.Write([]byte("welcome"))
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		fmt.Println("starting server on port " + httpPort)
		err := http.ListenAndServe(httpPort, r)
		if err != nil {
			panic(err)
		}

	}()

	go func() {
		defer wg.Done()
		fmt.Printf("starting job : %v\n", jobs[0].ID())
		s.Start()
	}()
	wg.Wait()

}

func initJobs(scheduler gocron.Scheduler) ([]gocron.Job, error) {
	// create a scheduler

	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("hello world")
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return []gocron.Job{j}, nil
}
func runCron() {

	//	fmt.Println(j.ID())

	// start the scheduler
	//	s.Start()

}
