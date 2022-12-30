package statics

import (
	"log"
	"testing"
)

func TestRequestStoreDetail(t *testing.T) {

	detail, err := RequestStoreDetail("1", "1", "473")

	if err != nil {
		log.Fatalln(err)
	}

	log.Println(detail)
}



