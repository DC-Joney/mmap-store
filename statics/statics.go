package statics

import (
	"strings"
	"sync/atomic"
)

var Statics *BookStatics = new(BookStatics)

type BookStatics struct {
	//内页图片数量
	pageImageCount int64

	//热区文本数量
	hotCount int64

	//热区多音频文本数量
	hotTextCount int64

	//内页多音频文本数量
	pageTextCount int64
}

func (statics *BookStatics) IncrementImageCount()  {
	atomic.AddInt64(&statics.pageImageCount, 1)
}

func (statics *BookStatics) IncrementHotCount()  {
	atomic.AddInt64(&statics.hotCount, 1)
}

func (statics *BookStatics) IncrementHotTextCount()  {
	atomic.AddInt64(&statics.hotTextCount, 1)
}

func (statics *BookStatics) AddHotTextCount(count int64)  {
	atomic.AddInt64(&statics.hotTextCount, count)
}

func (statics *BookStatics) IncrementPageTextCount()  {
	atomic.AddInt64(&statics.pageTextCount, 1)
}

func (statics *BookStatics) AddPageTextCount(count int64)  {
	atomic.AddInt64(&statics.pageTextCount, count)
}

func (statics *BookStatics) AddPagePiece(piece *BookPagePiece)  {

	//Logger.Infof("Piece message: %#v", piece)

	if strings.TrimSpace(piece.PieceContent) != "" {
		statics.IncrementHotCount()
	}

	for _,audio := range piece.Audios {
		if strings.TrimSpace(audio.Text) != "" {
			statics.IncrementHotTextCount()
		}
	}
}

func (statics *BookStatics) AddBookPage(page *BookPage)  {

	//Logger.Infof("Book audios size: %d", len(page.Audios))

	if strings.TrimSpace(page.ImageUrl) != "" {
		statics.IncrementImageCount()
	}

	for _,audio := range page.Audios {
		if strings.TrimSpace(audio.AudioText) != "" {
			statics.IncrementPageTextCount()
		}
	}

}



