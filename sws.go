package codec

import (
	/*
		#cgo linux,amd64 pkg-config: libav_linux_amd64.pc

		#include <stdio.h>
		#include <stdlib.h>
		#include <stdint.h>
		#include <string.h>
		#include "libavcodec/avcodec.h"
		#include "libavutil/avutil.h"
		#include "libavutil/opt.h"
		#include "libswscale/swscale.h"

		typedef struct {
			struct SwsContext *ctx;
		} swscale_t;

		static int init_swscale(swscale_t *m, int s_width, int s_height, int s_pixfmt,int d_width, int d_height, int d_pixfmt) {
			m->ctx = sws_getContext(s_width, s_height,
									s_pixfmt,
									d_width, d_height,
									s_pixfmt,
									SWS_BICUBIC, NULL, NULL, NULL);

			return 0;
		}

		static int av_sws_scale(swscale_t *m, AVFrame *src, int y, int h, AVFrame *dst) {
		 	av_log(m->ctx, AV_LOG_DEBUG, "Dst frame, w:%d, h%d, pix_fmt:%d\n",dst->width,dst->height,dst->format);
		 	av_log(m->ctx, AV_LOG_DEBUG, "Src frame, w:%d, h%d, pix_fmt:%d\n",src->width,src->height,src->format);

			int r = sws_scale(m->ctx, (const uint8_t *const*)src->data, src->linesize,
				y, h, (uint8_t *const*)dst->data, dst->linesize);

			return r;
		}


		static int sws_release(swscale_t *m) {
			sws_freeContext(m->ctx);
		}


	*/
	"C"
	"fmt"
	"image"
)

type AVSWScale struct {
	SrcWidth  int
	SrcHeight int
	SrcPixfmt image.YCbCrSubsampleRatio
	DstWidth  int
	DstHeight int
	DstPixfmt image.YCbCrSubsampleRatio
	Flags     int

	DstFrame *AVFrame

	m C.swscale_t
}

func (sws AVSWScale) Init() (*AVSWScale, error) {
	sPixFmt := pixFmtToAV(sws.SrcPixfmt)
	if sPixFmt == C.AV_PIX_FMT_NONE {
		return nil, fmt.Errorf("Src PIX_FMT invalid, %d", sws.SrcPixfmt)
	}

	dPixFmt := pixFmtToAV(sws.DstPixfmt)
	if sPixFmt == C.AV_PIX_FMT_NONE {
		return nil, fmt.Errorf("Dst PIX_FMT invalid, %d", sws.DstPixfmt)
	}

	var err error
	sws.DstFrame, err = CreateVideoFrame(sws.DstWidth, sws.DstHeight, sws.DstPixfmt)
	if err != nil {
		return nil, err
	}
	//log.Printf("Dst frame: %+v", sws.DstFrame)

	C.init_swscale(&sws.m, C.int(sws.SrcWidth), C.int(sws.SrcHeight), sPixFmt, C.int(sws.DstWidth), C.int(sws.DstHeight), dPixFmt)

	return &sws, nil
}

func (sws *AVSWScale) Scale(src *AVFrame) error {
	r := C.av_sws_scale(&sws.m, src.f, 0, C.int(sws.SrcHeight), sws.DstFrame.f)

	if r <= 0 {
		return fmt.Errorf("Scale error, %d", r)
	}

	return nil
}

func (sws *AVSWScale) Release() error {
	C.sws_release(&sws.m)
	sws.DstFrame.Release()

	return nil
}
