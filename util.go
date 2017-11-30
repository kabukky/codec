package codec

import (
	"image"
	"reflect"
	"unsafe"

	/*
		#cgo linux,amd64 pkg-config: libav_linux_amd64.pc

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

import (
	"sync"
)

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

func pixFmtToAV(pixFmt image.YCbCrSubsampleRatio) C.int {
	switch pixFmt {
	case image.YCbCrSubsampleRatio444:
		return C.AV_PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		return C.AV_PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		return C.AV_PIX_FMT_YUV420P
	}

	return C.AV_PIX_FMT_NONE
}
