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
			//av_log_set_level(AV_LOG_DEBUG);
			av_log_set_level(0);
		}
	*/
	"C"
)

import "sync"

const (
	AV_LOG_QUIET   = -8
	AV_LOG_PANIC   = 0
	AV_LOG_FATAL   = 8
	AV_LOG_ERROR   = 16
	AV_LOG_WARNING = 24
	AV_LOG_INFO    = 32
	AV_LOG_VERBOSE = 40
	AV_LOG_DEBUG   = 48
)

// open /close function not thread safe
var avLock sync.Mutex

func init() {
	//C.libav_init()

	C.av_register_all()
	SetLogLevel(AV_LOG_QUIET)
}

func SetLogLevel(level int) {
	C.av_log_set_level(C.int(level))
}

func fromCPtr(buf unsafe.Pointer, size int) (ret []uint8) {
	hdr := (*reflect.SliceHeader)((unsafe.Pointer(&ret)))
	hdr.Cap = size
	hdr.Len = size
	hdr.Data = uintptr(buf)
	return
}
