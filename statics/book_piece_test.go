package statics

import (
	"fmt"
	"log"
	"testing"
)

func TestRequestBookPagePiece(t *testing.T) {

	param := NewBookPagePieceParam(1, "2c05396f8b594620bbad5ec5861e7faf", "bc144ecd4b354f538dae302afc3856d4")

	pieces, err := RequestBookPagePiece(param)

	if err != nil {
		log.Fatalln(err)
	}

	for _,piece := range pieces {
		fmt.Printf("%#v\n", piece)
		fmt.Printf("%#v\n", piece.Audios[0].Text)
	}

}



