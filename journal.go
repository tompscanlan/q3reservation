package q3reservation

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type JournalEntry struct {
	Id      int    `json:"id"`
	Message string `json:"message"`
}

var JournalId = 0

func NewJournalEntry(msg string) *JournalEntry {

	// encode the string into a new journal entry
	entry := new(JournalEntry)
	entry.Message = base64.StdEncoding.EncodeToString([]byte(msg))
	entry.Id = JournalId
	JournalId += 1

	return entry
}

func PostJournalEntry(server string, entry JournalEntry) error {
	log.Println("sending post to journal")

	// add entry to the body
	jsonStr, err := json.Marshal(entry)
	if err != nil {
		log.Println(err)
		return err
	}

	// create the request
	url := fmt.Sprintf("%s/%s", server, "api/topic/10")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}

	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	log.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	log.Println("response Body:", string(body))
	return nil
}
