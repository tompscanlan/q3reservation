package main

import (
	//	"encoding/json"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/tompscanlan/labreserved/client"
	"github.com/tompscanlan/labreserved/models"

	"github.com/tompscanlan/labreserved/client/operations"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tompscanlan/q3reservation"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	verbose = kingpin.Flag("verbose", "Verbose mode.").Short('v').Bool()
	port    = kingpin.Flag("port", "port to listen on").Default(listenPortDefault).OverrideDefaultFromEnvar("PORT").Short('l').Int()

	journalServer = kingpin.Flag("journal-server", "REST endpoint for the journal server").Default(journalServerDefault).OverrideDefaultFromEnvar("JOURNAL_SERVER").Short('j').String()
	labDataServer = kingpin.Flag("labdata-server", "REST endpoint for the lab data server").Default(labDataServerDefault).OverrideDefaultFromEnvar("LABDATA_SERVER").Short('d').String()
	lock          = sync.RWMutex{}

	Client *apiclient.Labreserved
)

type Reservation struct {
	Name       string    `json:"name,omitempty"`
	StartDate  time.Time `json:"start_date,omitempty"`
	EndDate    time.Time `json:"end_date,omitempty"`
	ServerName string    `json:"server_name,omitempty"`
}

const (
	listenPortDefault    = "8082"
	journalServerDefault = "http://journal.butterhead.net:8080"
	labDataServerDefault = "labreserved.butterhead.net:2080"
)

func init() {
	setupFlags()
	q3reservation.Verbose = *verbose
}

func setupFlags() {
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()
}

func main() {

	transport := httptransport.New(*labDataServer, "", []string{"http"})
	Client = apiclient.New(transport, strfmt.Default)

	api := rest.NewApi()

	statusMw := &rest.StatusMiddleware{}
	api.Use(statusMw)
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Get("/.status", func(w rest.ResponseWriter, r *rest.Request) {
			w.WriteJson(statusMw.GetStatus())
		}),

		// whee... trailing slashes!
		// following existing api
		rest.Get("/api/servers/", GetAllServers),
		rest.Post("/api/servers/", PostServer),
		rest.Get("/api/reservations/", GetAllReservations),
		rest.Post("/api/reservations/", PostReservation),
		rest.Get("/api/updaters/:team", GetUpdater),
		rest.Post("/api/updaters/:team", PostUpdater),
	)
	if err != nil {
		log.Fatal(err)
	}
	api.SetApp(router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), api.MakeHandler()))
}

func GetAllServers(w rest.ResponseWriter, r *rest.Request) {
	params := operations.NewGetItemsParamsWithTimeout(30 * time.Second)
	resp, err := Client.Operations.GetItems(params)

	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(resp.Payload)
}
func PostServer(w rest.ResponseWriter, r *rest.Request) {

	params := operations.NewPostItemParamsWithTimeout(30 * time.Second)

	item := new(models.Item)
	err := r.DecodeJsonPayload(item)
	log.Printf("item: %s", item.String())
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params.Additem = item
	resp, err := Client.Operations.PostItem(params)

	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteJson(resp.Payload)

}

func GetAllReservations(w rest.ResponseWriter, r *rest.Request) {
	params := operations.NewGetItemsParamsWithTimeout(30 * time.Second)
	resp, err := Client.Operations.GetItems(params)

	var reservations []Reservation

	// each server
	for _, serv := range resp.Payload {
		// each reservation per server
		for _, reserv := range serv.Reservations {

			// convert to soapui API version of a reservation
			soapres := new(Reservation)

			soapres.Name = *serv.Name
			soapres.ServerName = *serv.Name
			soapres.StartDate = reserv.BeginTime()
			soapres.EndDate = reserv.EndTime()

			reservations = append(reservations, *soapres)
		}
	}
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(reservations)
}
func PostReservation(w rest.ResponseWriter, r *rest.Request) {

	params := operations.NewPostItemNameReservationParamsWithTimeout(30 * time.Second)

	reservation := new(models.Reservation)
	err := r.DecodeJsonPayload(reservation)
	log.Printf("item: %s", reservation.String())
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	params.Reservation = reservation
	resp, err := Client.Operations.PostItemNameReservation(params)

	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteJson(resp.Payload)

}

func GetUpdater(w rest.ResponseWriter, r *rest.Request) {
}
func PostUpdater(w rest.ResponseWriter, r *rest.Request) {
}
