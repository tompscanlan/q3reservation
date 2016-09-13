package q3reservation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tompscanlan/q3updater"
	"io/ioutil"
	"log"
	"net/http"
)

func GetUpdater(w rest.ResponseWriter, r *rest.Request) {

	url := fmt.Sprintf("%s/%s", UpdaterServer, "active")

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

	url := fmt.Sprintf("%s/%s", UpdaterServer, "active")
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
