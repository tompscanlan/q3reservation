package main

import (
	//	"encoding/json"
	"bytes"
	"encoding/json"
	"fmt"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	apiclient "github.com/tompscanlan/labreserved/client"
	"github.com/tompscanlan/labreserved/client/operations"
	"github.com/tompscanlan/labreserved/models"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tompscanlan/q3reservation"
	"github.com/tompscanlan/q3updater"

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
	updaterServerDefault = "http://127.0.0.1:8083"
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
		rest.Put("/api/updaters/:team", PutUpdater),
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

	url := fmt.Sprintf("%s/%s", *updaterServer, "active")

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {

		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))

	active := q3updater.NewActive()
	err = json.Unmarshal(body, active)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteJson(active)
}
func PutUpdater(w rest.ResponseWriter, r *rest.Request) {
	active := new(q3updater.Active)

	err := r.DecodeJsonPayload(active)
	if err != nil {
		log.Println("parsing input body: ", err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	log.Printf("got in put active: %s", active.String())

	var body []byte

	jsonStr, err := json.Marshal(active)
	if err != nil {

		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url := fmt.Sprintf("%s/%s", *updaterServer, "active")
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonStr))

	if err != nil {

		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		rest.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ = ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	return
}
