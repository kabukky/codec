package codec

import (
	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -L/usr/local/lib  -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

		#include "libavcodec/avcodec.h"
		#include "libavutil/avutil.h"
		#include <string.h>
		#include <stdio.h>

		typedef struct {
			AVCodec *c;
			AVCodecContext *ctx;
			AVFrame *f;
			int got;
		} aacdec_t ;

		static int aacdec_new(aacdec_t *m, int codec_id, uint8_t *buf, int len) {
			m->c = avcodec_find_decoder(codec_id);
			m->ctx = avcodec_alloc_context3(m->c);
			m->f = av_frame_alloc();
			if(len > 0 && buf != NULL) {
				m->ctx->extradata = buf;
				m->ctx->extradata_size = len;
			}
			m->ctx->debug = 0x3;
			m->ctx->channels = 1;
			m->ctx->channel_layout = av_get_default_channel_layout(1);
			m->ctx->sample_fmt = m->c->sample_fmts[0];

			av_log(m->ctx, AV_LOG_DEBUG, "m %p, ch_layout: %ld, sample_fmt: %d\n", m,m->ctx->channel_layout,m->ctx->sample_fmt);

			int r = avcodec_open2(m->ctx, m->c, 0);
			if(r != 0) {
				static char error_buffer[255];
				av_strerror(r, error_buffer, sizeof(error_buffer));
				av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
			}
			return r;
		}

		static int aacdec_decode(aacdec_t *m, uint8_t *data, int len) {
			AVPacket pkt;
			av_init_packet(&pkt);
			pkt.data = data;
			pkt.size = len;

			int r = avcodec_decode_audio4(m->ctx, m->f, &m->got, &pkt);
			if(r < 0) {
				static char error_buffer[255];
				av_strerror(r, error_buffer, sizeof(error_buffer));
				av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
			}
			if(r > 0 && len > r) {
				av_log(m->ctx, AV_LOG_DEBUG, "return positive result %d\n", r);
			}

			return r;
		}
	*/
	"C"
	"errors"
	"unsafe"
)

type AACDecoder struct {
	m C.aacdec_t
}

func NewAACDecoder(codec string, header []byte) (m *AACDecoder, err error) {
	m = &AACDecoder{}

	codec_id := 0
	switch codec {
	case "aac":
		codec_id = C.AV_CODEC_ID_AAC
	case "mulaw":
		codec_id = C.AV_CODEC_ID_PCM_MULAW
	case "alaw":
		codec_id = C.AV_CODEC_ID_PCM_ALAW
	}

	var r C.int

	avLock.Lock()
	defer avLock.Unlock()

	if header == nil {
		r = C.aacdec_new(&m.m, C.int(codec_id), (*C.uint8_t)(unsafe.Pointer(nil)), (C.int)(0))
	} else {
		r = C.aacdec_new(&m.m,
			C.int(codec_id),
			(*C.uint8_t)(unsafe.Pointer(&header[0])),
			(C.int)(len(header)),
		)
	}

	if int(r) < 0 {
		err = errors.New("open codec failed")
	}

	return
}

func (m *AACDecoder) Decode(data []byte) (sample []byte, err error) {
	r := C.aacdec_decode(
		&m.m,
		(*C.uint8_t)(unsafe.Pointer(&data[0])),
		(C.int)(len(data)),
	)
	if int(r) < 0 {
		err = errors.New("decode failed")
		return
	}
	if int(m.m.got) == 0 {
		err = errors.New("no data")
		return
	}

	var channels = 1
	//size := int(m.m.f.linesize[0]) * channels
	size := int(m.m.f.nb_samples*2) * channels
	sample = make([]byte, size*channels)
	for i := 0; i < channels; i++ {
		C.memcpy(
			unsafe.Pointer(&sample[i*size]),
			unsafe.Pointer(m.m.f.data[i]),
			(C.size_t)(size),
		)
	}
	return
}
