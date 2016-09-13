package q3reservation

import (
	"encoding/base64"
	"encoding/json"
	"log"
	"time"
)

// Verbose lets other packages know we want to be noisey
var (
	Verbose       = true
	UpdaterServer = ""
	JournalServer = ""
)

type SoapReservation struct {
	Name       string    `json:"name,omitempty"`
	StartDate  time.Time `json:"start_date,omitempty"`
	EndDate    time.Time `json:"end_date,omitempty"`
	ServerName string    `json:"server_name,omitempty"`
	Approved   bool      `json:"approved"`
}

func NewReservation() *SoapReservation {
	res := new(SoapReservation)
	return res
}
func (r SoapReservation) String() string {
	b := r.ByteArray()
	return string(b[:])
}

func (r *SoapReservation) FromJson(bytes []byte) *SoapReservation {
	if err := json.Unmarshal(bytes, r); err != nil {
		log.Println(err)
	}
	return r
}

func (r SoapReservation) ByteArray() []byte {
	b, err := json.Marshal(r)
	if err != nil {
		log.Println(err)
	}
	return b
}

// take reservation, convert to json string, base 64 encode it
// and put it in the message of a journal entry
func (r *SoapReservation) ToJournalEntry() *JournalEntry {
	entry := new(JournalEntry)

	b := r.ByteArray()

	entry.Message = base64.StdEncoding.EncodeToString(b)
	entry.Id = JournalId
	JournalId += 1

	return entry
}

func (r *SoapReservation) FromJournalEntry(je JournalEntry) *SoapReservation {
	decoded, err := base64.StdEncoding.DecodeString(je.Message)
	if err != nil {
		log.Println(err)
	}
	return r.FromJson(decoded)

}
