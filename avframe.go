package codec

import (
	/*
		#cgo CFLAGS:  -I/usr/local/include
		#cgo LDFLAGS: -L/usr/local/lib -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

		#include "libavcodec/avcodec.h"
		#include "libavutil/avutil.h"
		#include "libavformat/avformat.h"
	*/
	"C"
)
import (
	"fmt"
	"image"
	"unsafe"
)

type AVFrame struct {
	f *C.AVFrame
}

func NewAVFrame() *AVFrame {
	return &AVFrame{C.av_frame_alloc()}
}

func (f *AVFrame) Release() {
	C.av_frame_free(&f.f)
}

func (f *AVFrame) GetPktPts() int64 {
	return int64(f.f.pkt_pts)
}

func (f *AVFrame) GetPktDts() int64 {
	return int64(f.f.pkt_dts)
}

func (f *AVFrame) ImgCopy(img *image.YCbCr) error {
	if int(f.f.width) != img.Rect.Max.X || int(f.f.height) != img.Rect.Max.Y {
		return fmt.Errorf("Decode2: invalid image size, %dx%d  vs %dx%d", int(f.f.width), int(f.f.height), img.Rect.Max.X, img.Rect.Max.Y)
	}

	img.YStride = int(f.f.linesize[0])
	img.CStride = int(f.f.linesize[1])

	C.memcpy(
		unsafe.Pointer(&img.Y[0]),
		unsafe.Pointer(f.f.data[0]),
		(C.size_t)(img.YStride*img.Rect.Max.Y),
	)
	C.memcpy(
		unsafe.Pointer(&img.Cb[0]),
		unsafe.Pointer(f.f.data[1]),
		(C.size_t)(img.CStride*img.Rect.Max.Y/2),
	)
	C.memcpy(
		unsafe.Pointer(&img.Cr[0]),
		unsafe.Pointer(f.f.data[2]),
		(C.size_t)(img.CStride*img.Rect.Max.Y/2),
	)

	return nil
}

func (f *AVFrame) ImgBind(img *image.YCbCr) error {
	if int(f.f.width) != img.Rect.Max.X || int(f.f.height) != img.Rect.Max.Y {
		return fmt.Errorf("Decode2: invalid image size, %dx%d  vs %dx%d", int(f.f.width), int(f.f.height), img.Rect.Max.X, img.Rect.Max.Y)
	}

	img.YStride = int(f.f.linesize[0])
	img.CStride = int(f.f.linesize[1])

	img.Y = fromCPtr(unsafe.Pointer(f.f.data[0]), img.YStride*img.Rect.Max.Y)
	img.Cb = fromCPtr(unsafe.Pointer(f.f.data[1]), img.CStride*img.Rect.Max.Y/2)
	img.Cr = fromCPtr(unsafe.Pointer(f.f.data[2]), img.CStride*img.Rect.Max.Y/2)

	return nil
}
