package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/S1FFFkA/PET_PROJECT_N1_WEATHER/internal/client/HTTP/geocoding"
	open_meteo "github.com/S1FFFkA/PET_PROJECT_N1_WEATHER/internal/client/HTTP/open-meteo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
)

const httpPort = ":3000"
const city = "moscow"

type Reading struct {
	Name        string    `db:"name"`
	Timestamp   time.Time `db:"timestamp"`
	Temperature float64   `db:"temperature"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgresql://gleb:123@localhost:54321/weather")
	if err != nil {
		panic(err)
	}
	defer conn.Close(context.Background())
	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {

		cityName := chi.URLParam(r, "city")
		fmt.Fprintf(w, "Hello, %s\n!", cityName)
		var reading Reading
		err = conn.QueryRow(
			ctx,
			"select name ,timestamp,temperature from reading where name = $1 order by timestamp desc limit 1;", city,
		).Scan(&reading.Name, &reading.Timestamp, &reading.Temperature)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
		}

		var raw []byte
		raw, err = json.Marshal(reading)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		_, err = w.Write(raw)
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}
	jobs, err := initJobs(ctx, s, conn)
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

func initJobs(ctx context.Context, scheduler gocron.Scheduler, conn *pgx.Conn) ([]gocron.Job, error) {
	// create a scheduler
	httpClient := &http.Client{Timeout: time.Second * 10}
	geocodingClient := geocoding.NewClient(httpClient)
	openMeteoClient := open_meteo.NewClient(httpClient)

	j, err := scheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				geores, err := geocodingClient.GetCoords(city)
				if err != nil {
					log.Println(err)
					return
				}

				openMetRes, err := openMeteoClient.GetTemperature(geores.Latitude, geores.Longitude)
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("Raw time from API: %s, Temperature: %.1f", openMetRes.Current.Time, openMetRes.Current.Temperature2m)
				timestamp, err := time.Parse("2006-01-02T15:04", openMetRes.Current.Time)
				if err != nil {
					log.Println(err)
					return
				}
				log.Printf("Parsed timestamp: %s", timestamp.Format("2006-01-02 15:04:05 MST"))
				_, err = conn.Exec(ctx, "insert into reading (name, temperature ,timestamp ) values ($1,$2,$3)",
					city, openMetRes.Current.Temperature2m, timestamp)
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Printf("%vupdated data fo city  %s\n", time.Now(), city)
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
