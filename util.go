/*

Golang h264,aac decoder/encoder libav wrapper

	d, err = codec.NewAACEncoder()
	data, err = d.Encode(samples)

	d, err = codec.NewAACDecoder(aaccfg)
	samples, err = d.Decode(data)

	var img *image.YCbCr
	d, err = codec.NewH264Encoder(640, 480)
	img, err = d.Encode(img)

	d, err = codec.NewH264Decoder(pps)
	img, err = d.Decode(nal)
*/
package codec

import (
	"reflect"
	"unsafe"

	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

		#include "libavutil/avutil.h"
		#include "libavformat/avformat.h"

		static void libav_init() {
			av_register_all();
			av_log_set_level(AV_LOG_DEBUG);
		}
	*/
	"C"
)

import "sync"

// open /close function not thread safe
var avLock sync.Mutex

func init() {
	C.libav_init()
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)((unsafe.Pointer(&ret)))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}
