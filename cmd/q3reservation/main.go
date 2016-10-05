package main

import (
	//	"encoding/json"
	"encoding/json"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/tompscanlan/labreserved/client"
	"github.com/tompscanlan/labreserved/client/operations"
	"github.com/tompscanlan/labreserved/models"
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
	updaterServer = kingpin.Flag("updater-server", "REST endpoint for the updater server").Default(updaterServerDefault).OverrideDefaultFromEnvar("UPDATER_SERVER").Short('u').String()
	lock          = sync.RWMutex{}

	Client *apiclient.Labreserved
)

const (
	listenPortDefault    = "8082"
	journalServerDefault = "http://journal.butterhead.net:8080"
	labDataServerDefault = "labreserved.butterhead.net:2080"
	updaterServerDefault = "http://127.0.0.1:8083"
)

func init() {
	setupFlags()

	q3reservation.Verbose = *verbose
	q3reservation.UpdaterServer = *updaterServer
	q3reservation.JournalServer = *journalServer
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
		rest.Get("/api/updaters/:team", q3reservation.GetUpdater),
		rest.Put("/api/updaters/:team", q3reservation.PutUpdater),
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
	log.Println("get all reservations")

	// the list of all reservations from the lab server
	var reservations []q3reservation.SoapReservation

	// get reservations from the lab data server
	params := operations.NewGetItemsParamsWithTimeout(30 * time.Second)
	resp, err := Client.Operations.GetItems(params)

	if err != nil {
		log.Println("GetAllReservations:GetItems error:", err)
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	log.Println("data from lab service: ", resp.Payload)

	// each server
	for _, serv := range resp.Payload {

		log.Printf("doing server %s", *serv.Name)
		// each reservation per server
		for _, reserv := range serv.Reservations {

			// convert to soapui API version of a reservation
			soapres := new(q3reservation.SoapReservation)

			//			soapres.Name = *serv.Name
			//			soapres.ServerName = *serv.Name
			soapres.Name = *reserv.Username
			soapres.ServerName = *serv.Name
			err, soapres.StartDate = reserv.GetTime()
			err, soapres.EndDate = reserv.GetEndTime()
			soapres.Approved = *reserv.Approved

			reservations = append(reservations, *soapres)
			log.Printf("found reservations %s", reservations)

		}
	}
	if err != nil {
		rest.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.WriteJson(reservations)
}

func PostReservation(w rest.ResponseWriter, r *rest.Request) {

	// pull reservation from the request
	reservation := new(q3reservation.SoapReservation)
	err := r.DecodeJsonPayload(reservation)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("received reservation: %s", reservation.String())

	jsonStr, err := json.Marshal(reservation)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	entry := q3reservation.NewJournalEntry(string(jsonStr[:]))
	err = q3reservation.PostJournalEntry(*journalServer, *entry)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// also write the request back to the browser
	w.WriteJson(reservation)
}
