package codec

import (
	/*
		#cgo linux,amd64 pkg-config: libav_linux_amd64.pc
		#include "libavcodec/avcodec.h"
		#include "libavutil/avutil.h"
		#include "libavutil/pixfmt.h"
		#include "libavformat/avformat.h"

		static int create_video_frame(AVFrame **pf, int width, int height, int pix_fmt) {
			AVFrame *f;

			f = av_frame_alloc();

		    f->format = pix_fmt;
		    f->width  = width;
		    f->height = height;

		 	av_log(NULL, AV_LOG_DEBUG, "Allocate AVFrame, w:%d, h%d, pix_fmt:%d\n",f->width,f->height,f->format);

			if (av_frame_get_buffer(f, 32) < 0) {
				 av_log(NULL, AV_LOG_DEBUG, "Could not allocate frame data.\n");

				 return 1;
			}

			*pf = f;

			return 0;
		}

		static int is_refcounted_video_frame(AVFrame *f) {
			if(f->buf == NULL && f->extended_buf == NULL) {
				return 0;
			}

			return 1;
		}

		static int release_video_frame(AVFrame **pf) {
			if((*pf)->buf != NULL || (*pf)->extended_buf != NULL) {
				av_frame_unref(*pf);
			}

			av_frame_free(pf);
		}
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

func CreateVideoFrame(width, height int, pixFmt image.YCbCrSubsampleRatio) (*AVFrame, error) {
	f := &AVFrame{}

	//avPixFmt := pixFmtToAV(pixFmt)
	//log.Printf("Create Video frame, pixFmt:%+v, avPixFmt:%d", pixFmt, avPixFmt)
	r := C.create_video_frame(&f.f, C.int(width), C.int(height), pixFmtToAV(pixFmt))
	if r > 0 {
		return nil, fmt.Errorf("Create AVFrame failed")
	}

	return f, nil
}

func (f *AVFrame) Release() {
	// r := C.is_refcounted_video_frame(f.f)
	// if int(r) == 1 {
	// 	log.Println("Relese REFCOUNTED frame")
	// }

	C.release_video_frame(&f.f)
	//C.av_frame_free(&f.f)
}

func (f *AVFrame) GetPktPts() int64 {
	return int64(f.f.pkt_pts)
}

func (f *AVFrame) GetPktDts() int64 {
	return int64(f.f.pkt_dts)
}

func (f *AVFrame) GetSize() image.Rectangle {
	return image.Rectangle{
		Min: image.Point{0, 0},
		Max: image.Point{int(f.f.width), int(f.f.height)},
	}
}

// ImgCopy1 - Copy frame to image.YCbCr
// In AVFrame data planes may be greater, then image planes, libav use padding
// If plane sizes equal, copy all plane at once
// Else copy plane line by line
func (f *AVFrame) ImgCopy1(img *image.YCbCr) error {
	ir := img.Rect.Max
	fr := image.Point{int(f.f.width), int(f.f.height)}

	if fr.X != ir.X || fr.Y != ir.Y {
		return fmt.Errorf("AVFrame, image copy: invalid image size, fr:%+v  vs ir:%+v", fr, ir)
	}

	// 420P in avlib -> 0
	if f.f.format != 0 {
		return fmt.Errorf("Invalid PIX_FMT, %d", f.f.format)
	}

	if img.YStride == int(f.f.linesize[0]) {
		C.memcpy(
			unsafe.Pointer(&img.Y[0]),
			unsafe.Pointer(f.f.data[0]),
			(C.size_t)(img.YStride*ir.Y),
		)
	} else {
		for i := 0; i < ir.Y; i++ {
			dst := unsafe.Pointer(&img.Y[i*img.YStride])
			src := unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[0])) + uintptr(i*int(f.f.linesize[0])))

			C.memcpy(dst, src, (C.size_t)(ir.X))
		}
	}

	if img.CStride == int(f.f.linesize[1]) {
		C.memcpy(
			unsafe.Pointer(&img.Cb[0]),
			unsafe.Pointer(f.f.data[1]),
			(C.size_t)(img.CStride*ir.Y/2),
		)

		C.memcpy(
			unsafe.Pointer(&img.Cr[0]),
			unsafe.Pointer(f.f.data[2]),
			(C.size_t)(img.CStride*ir.Y/2),
		)
	} else {
		for i := 0; i < ir.Y/2; i++ {
			dst := unsafe.Pointer(&img.Cb[i*img.CStride])
			src := unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[1])) + uintptr(i*int(f.f.linesize[1])))
			C.memcpy(dst, src, (C.size_t)(ir.X/2))

			dst = unsafe.Pointer(&img.Cr[i*img.CStride])
			src = unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[2])) + uintptr(i*int(f.f.linesize[1])))
			C.memcpy(dst, src, (C.size_t)(ir.X/2))
		}
	}

	return nil
}

// ImgFit - Fit frame in image
func (f *AVFrame) ImgFit(img *image.YCbCr) error {
	if f.f.format != 0 {
		return fmt.Errorf("Invalid PIX_FMT, %d", f.f.format)
	}

	h := img.Rect.Max.Y
	hs := 0
	if h > int(f.f.height) {
		hs = (h - int(f.f.height)) / 2
		h = int(f.f.height)

		//log.Fatalf("Arrange image, img_h:%d, frame_h:%d, hs:%d", img.Rect.Max.Y, int(f.f.height), hs)
	}

	w := img.Rect.Max.X
	ws := 0
	if w > int(f.f.width) {
		ws = (w - int(f.f.width)) / 2
		w = int(f.f.width)
	}

	// clean Y plane, image reuse
	C.memset(unsafe.Pointer(&img.Y[0]), 0, (C.size_t)(img.YStride*img.Rect.Max.Y))

	if hs == 0 && ws == 0 && img.YStride == int(f.f.linesize[0]) {
		C.memcpy(
			unsafe.Pointer(&img.Y[0]),
			unsafe.Pointer(f.f.data[0]),
			(C.size_t)(img.YStride*h),
		)
	} else {
		for i := 0; i < h; i++ {
			dst := unsafe.Pointer(&img.Y[(i+hs)*img.YStride+ws])
			src := unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[0])) + uintptr(i*int(f.f.linesize[0])))

			C.memcpy(dst, src, (C.size_t)(w))
		}
	}

	// clean Cb,Cr planes, image reuse
	C.memset(
		unsafe.Pointer(&img.Cb[0]),
		128,
		(C.size_t)(img.CStride*img.Rect.Max.Y/2),
	)
	C.memset(
		unsafe.Pointer(&img.Cr[0]),
		128,
		(C.size_t)(img.CStride*img.Rect.Max.Y/2),
	)

	if hs == 0 && ws == 0 && img.CStride == int(f.f.linesize[1]) {
		C.memcpy(
			unsafe.Pointer(&img.Cb[0]),
			unsafe.Pointer(f.f.data[1]),
			(C.size_t)(img.CStride*h/2),
		)

		C.memcpy(
			unsafe.Pointer(&img.Cr[0]),
			unsafe.Pointer(f.f.data[2]),
			(C.size_t)(img.CStride*h/2),
		)
	} else {
		for i := 0; i < h/2; i++ {
			dst := unsafe.Pointer(&img.Cb[(i+hs/2)*img.CStride+ws/2])
			src := unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[1])) + uintptr(i*int(f.f.linesize[1])))
			C.memcpy(dst, src, (C.size_t)(w/2))

			dst = unsafe.Pointer(&img.Cr[(i+hs/2)*img.CStride+ws/2])
			src = unsafe.Pointer(uintptr(unsafe.Pointer(f.f.data[2])) + uintptr(i*int(f.f.linesize[1])))
			C.memcpy(dst, src, (C.size_t)(w/2))
		}
	}

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
