package statics

import (
	"log"
	"testing"
)

func TestRequestBookPage(t *testing.T) {

	//{"apiKey":"9f6fdcedf5c44111b5e14d632c043189","bookSign":"8ad720950e4e4c88aa3d93dd0e68bea3","pageSize":6,"pageNo":1}

	param := NewBookPageParam(6, "2c05396f8b594620bbad5ec5861e7faf", "bc144ecd4b354f538dae302afc3856d4")

	pages, err := RequestBookPage(param)

	if err != nil {
		log.Fatalln("ErrorMessage: ", err)
	}

	log.Printf("%#v",pages)
}
