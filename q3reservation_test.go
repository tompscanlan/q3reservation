package q3reservation

import (
	"testing"
	"time"
)

var resTests = []struct {
	in    SoapReservation
	valid bool
}{
	{
		SoapReservation{Name: "testing", StartDate: time.Now(), EndDate: time.Now().Add(3 * time.Hour), ServerName: "someserver"},
		true,
	},
	{
		SoapReservation{Name: "testing", StartDate: time.Now().Add(6 * time.Hour), EndDate: time.Now().Add(24 * time.Hour), ServerName: "someserver-1"},
		true,
	},
}

func CompareRes(a *SoapReservation, b *SoapReservation, t *testing.T) {

	if a.Approved != b.Approved {
		t.Errorf("expected %t, got %t", a.Approved, b.Approved)
	}

	if a.Name != b.Name {
		t.Errorf("expected %s, got %s", a.Name, b.Name)
	}

	if a.ServerName != b.ServerName {
		t.Errorf("expected %s, got %s", a.ServerName, b.ServerName)
	}
	if !a.StartDate.Equal(b.StartDate) {
		t.Errorf("expected %s, got %s", a.StartDate, b.StartDate)
	}

	if !a.EndDate.Equal(b.EndDate) {
		t.Errorf("expected %s, got %s", a.EndDate, b.EndDate)
	}

}
func TestReservationToFromJSON(t *testing.T) {

	for _, test := range resTests {
		res := &test.in

		asJson := res.ByteArray()
		asRes := NewReservation().FromJson(asJson)

		CompareRes(res, asRes, t)

	}
}

func TestReservationToFromJournal(t *testing.T) {
	for _, test := range resTests {
		res := &test.in
		je := res.ToJournalEntry()

		newRes := NewReservation().FromJournalEntry(*je)

		CompareRes(res, newRes, t)
	}
}
