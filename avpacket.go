package codec

import (
	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

		#include "libavcodec/avcodec.h"
	*/
	"C"
)

type AVPacket struct {
	Data   []byte
	Pts    int64
	Dts    int64
	Key    bool
	AVFree bool

	pkt C.AVPacket
}

func NewAVPacket() *AVPacket {
	pkt := &AVPacket{}
	C.av_init_packet(&pkt.pkt)

	return pkt
}

func (pkt *AVPacket) Free() {
	if pkt.AVFree {
		C.av_free_packet(&pkt.pkt)
	}
}
