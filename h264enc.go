package codec

import (

	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

		#include <stdio.h>
		#include <stdlib.h>
		#include <stdint.h>
		#include <string.h>
		#include "libavcodec/avcodec.h"
		#include "libavutil/avutil.h"
		#include "libavformat/avformat.h"


		typedef struct {
			int w, h;
			int pixfmt;
			int64_t ppts;
			char *preset[2];
			char *profile;
			int bitrate;
			int framerate;
			int got;
			AVCodec *c;
			AVCodecContext *ctx;
			AVFrame *f;
			AVPacket pkt;
		} h264enc_t;

		static int h264enc_new(h264enc_t *m) {
			m->c = avcodec_find_encoder(AV_CODEC_ID_H264);
			m->ctx = avcodec_alloc_context3(m->c);
			m->ctx->width = m->w;
			m->ctx->height = m->h;
			m->ctx->bit_rate = m->bitrate;
			m->ctx->time_base = (AVRational){1,m->framerate};
			m->ctx->gop_size = 15;
			m->ctx->pix_fmt = m->pixfmt;
			m->ctx->flags |= CODEC_FLAG_GLOBAL_HEADER;

		    AVFrame *picture;
		    picture = av_frame_alloc();
		    picture->format = m->ctx->pix_fmt;
		    picture->width  = m->ctx->width;
		    picture->height = m->ctx->height;
		    if (av_frame_get_buffer(picture, 32) < 0) {
		        fprintf(stderr, "Could not allocate frame data.\n");
		    }
			av_frame_make_writable(picture);
			m->f = picture;

			return avcodec_open2(m->ctx, NULL, NULL);
		}

		static void set_ppts(h264enc_t *m, int64_t ppts) {
			m->ppts = ppts;
			m->f->pts = ppts;
		}

	*/
	"C"
	"errors"
	"image"
	"log"
	"strings"
	"unsafe"
)

type H264Encoder struct {
	m         C.h264enc_t
	Header    []byte
	Pixfmt    image.YCbCrSubsampleRatio
	W, H      int
	pts       int64
	FrameRate int
	Preset    [2]string
	Profile   string
}

func NewH264Encoder(
	w, h, frameRate int,
	pixfmt image.YCbCrSubsampleRatio,
	opts ...string,
) (m *H264Encoder, err error) {
	m = &H264Encoder{
		W:         w,
		H:         h,
		Pixfmt:    pixfmt,
		FrameRate: frameRate,
	}

	m.m.w = (C.int)(m.W)
	m.m.h = (C.int)(m.H)
	m.m.framerate = (C.int)(m.FrameRate)

	switch pixfmt {
	case image.YCbCrSubsampleRatio444:
		m.m.pixfmt = C.PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		m.m.pixfmt = C.PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		m.m.pixfmt = C.PIX_FMT_YUV420P
	}

	for _, opt := range opts {
		a := strings.Split(opt, ",")
		switch {
		case a[0] == "preset" && len(a) == 3:
			m.m.preset[0] = C.CString(a[1])
			m.m.preset[1] = C.CString(a[2])
		case a[0] == "profile" && len(a) == 2:
			m.m.profile = C.CString(a[1])
		}
	}

	avLock.Lock()
	r := C.h264enc_new(&m.m)
	avLock.Unlock()
	if int(r) < 0 {
		err = errors.New("open encoder failed")
		return
	}

	m.Header = fromCPtr(unsafe.Pointer(m.m.ctx.extradata), (int)(m.m.ctx.extradata_size))
	//m.Header = fromCPtr(unsafe.Pointer(m.m.pps), (int)(m.m.ppslen))
	return
}

type H264Out struct {
	pkt  C.AVPacket
	Data []byte
	Key  bool
}

func (out *H264Out) Free() {
	C.av_free_packet(&out.pkt)
}

func (m *H264Encoder) Init() error {
	//w, h, frameRate
	m.m.w = (C.int)(m.W)
	m.m.h = (C.int)(m.H)
	m.m.framerate = (C.int)(m.FrameRate)

	// set pixFmt
	switch m.Pixfmt {
	case image.YCbCrSubsampleRatio444:
		m.m.pixfmt = C.PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		m.m.pixfmt = C.PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		m.m.pixfmt = C.PIX_FMT_YUV420P
	}

	// set preset
	if len(m.Preset[0]) > 0 {
		m.m.preset[0] = C.CString(m.Preset[0])
	}
	if len(m.Preset[1]) > 0 {
		m.m.preset[1] = C.CString(m.Preset[1])
	}

	// set profile
	if len(m.Profile) > 0 {
		m.m.profile = C.CString(m.Profile)
	}

	avLock.Lock()
	defer avLock.Unlock()

	r := C.h264enc_new(&m.m)
	if int(r) < 0 {
		return errors.New("open encoder failed")
	}

	m.Header = fromCPtr(unsafe.Pointer(m.m.ctx.extradata), (int)(m.m.ctx.extradata_size))
	//m.Header = fromCPtr(unsafe.Pointer(m.m.pps), (int)(m.m.ppslen))

	return nil
}

func (m *H264Encoder) Encode(img *image.YCbCr) (out *H264Out, err error) {
	var f *C.AVFrame
	if img == nil {
		f = nil
	} else {
		if img.SubsampleRatio != m.Pixfmt {
			err = errors.New("image pixfmt not match")
			return
		}
		if img.Rect.Dx() != m.W || img.Rect.Dy() != m.H {
			err = errors.New("image size not match")
			return
		}
		f = m.m.f
		f.data[0] = (*C.uint8_t)(unsafe.Pointer(&img.Y[0]))
		f.data[1] = (*C.uint8_t)(unsafe.Pointer(&img.Cb[0]))
		f.data[2] = (*C.uint8_t)(unsafe.Pointer(&img.Cr[0]))
		f.linesize[0] = (C.int)(img.YStride)
		f.linesize[1] = (C.int)(img.CStride)
		f.linesize[2] = (C.int)(img.CStride)

		log.Println("avf pts:", m.pts)
		f.pts = (C.int64_t)(m.pts)
		C.set_ppts(&m.m, (C.int64_t)(m.pts))
		m.pts++
	}

	out = &H264Out{}

	C.av_init_packet(&out.pkt)
	out.pkt.data = nil
	out.pkt.size = 0

	r := C.avcodec_encode_video2(m.m.ctx, &out.pkt, f, &m.m.got)
	//defer C.av_free_packet(&m.m.pkt)
	if int(r) < 0 {
		err = errors.New("encode failed")
		return
	}
	if m.m.got == 0 {
		err = errors.New("no picture")
		return
	}
	if out.pkt.size == 0 {
		err = errors.New("packet size == 0")
		return
	}
	log.Println("pkt pts:", out.pkt.pts)
	// out.Data = make([]byte, m.m.pkt.size)
	// C.memcpy(
	// 	unsafe.Pointer(&out.Data[0]),
	// 	unsafe.Pointer(m.m.pkt.data),
	// 	(C.size_t)(m.m.pkt.size),
	// )
	out.Key = (out.pkt.flags & C.AV_PKT_FLAG_KEY) != 0

	return
}
