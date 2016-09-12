package main

import (
	"log"
	"testing"
)

func TestBlobSet(t *testing.T) {

	if testing.Short() {
		return
	}

	GetAllServers
	err := PostBlob(testBlobId, testStr)
	if err != nil {
		t.Error(err)
	}

	blob, err := GetBlob(testBlobId)
	if err != nil {
		t.Error(err)
	}
	if testing.Verbose() {
		log.Println("Got decoded blob content: ", string(blob))
	}

	if blob != testStr {
		t.Errorf("blob(%s) doesn't match (%s)", blob, testStr)
	}

}
