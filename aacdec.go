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
			uint8_t *out_buffer;
			int out_sz;
			int sample_size;
		} aacdec_t ;

		static int get_frame_buffer(AVCodecContext *ctx, AVFrame *frame, int flags) {
			av_log(NULL, AV_LOG_DEBUG, "Get output buffer for audio samples, nb_samples: %d\n", frame->nb_samples);
		}

		static int aacdec_new(aacdec_t *m, int codec_id, uint8_t *buf, int len) {
			m->c = avcodec_find_decoder(codec_id);
			m->ctx = avcodec_alloc_context3(m->c);
			m->f = av_frame_alloc();
			if(len > 0 && buf != NULL) {
				m->ctx->extradata = buf;
				m->ctx->extradata_size = len;
			}
			m->ctx->channels = 1;
			m->ctx->channel_layout = av_get_default_channel_layout(1);
			m->ctx->sample_fmt = m->c->sample_fmts[0];

			int r = avcodec_open2(m->ctx, m->c, 0);
			if(r != 0) {
				static char error_buffer[255];
				av_strerror(r, error_buffer, sizeof(error_buffer));
				av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
			}

			m->sample_size = av_get_bytes_per_sample(m->ctx->sample_fmt);

			av_log(m->ctx, AV_LOG_DEBUG, "Audio decoder:, channels: %d, ch_layout: %ld, sample_fmt: %d, planar: %d\n", m->ctx->channels,m->ctx->channel_layout,m->ctx->sample_fmt,av_sample_fmt_is_planar(m->ctx->sample_fmt));

			return r;
		}

		static void aacdec_release(aacdec_t *m) {
			// release context
			avcodec_close(m->ctx);
			av_free(m->ctx);

			// release frame
			av_frame_free(&m->f);
		}

		static int aacdec_decode(aacdec_t *m, uint8_t *in_data, int in_sz) {
			int i,linesize;

			AVPacket pkt;
			av_init_packet(&pkt);
			pkt.data = in_data;
			pkt.size = in_sz;

			int r = avcodec_decode_audio4(m->ctx, m->f, &m->got, &pkt);
			if(r < 0) {
				static char error_buffer[255];
				av_strerror(r, error_buffer, sizeof(error_buffer));
				av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
			}
			if(r > 0 && in_sz > r) {
				av_log(m->ctx, AV_LOG_DEBUG, "return positive result %d\n", r);
			}

			av_log(m->ctx, AV_LOG_DEBUG, "aac_decode, channels layout: %lu, channels: %d, nb_samples: %d, line size: %d\n", m->f->channel_layout, m->ctx->channels, m->f->nb_samples, m->f->linesize[0]);

			return r;
		}

		static int aacdec_decode2(aacdec_t *m, uint8_t *out_data, int out_sz, uint8_t *in_data, int in_sz) {
			int i,linesize;

			av_log(m->ctx, AV_LOG_DEBUG, "audio decode2\n");
			//return aacdec_decode(m, in_data, in_sz);

			AVPacket pkt;
			av_init_packet(&pkt);
			pkt.data = in_data;
			pkt.size = in_sz;

			linesize = out_sz / m->ctx->channels;
			for (i=0; i<m->ctx->channels; ++i) {
				m->f->data[i] = &out_data[i*linesize];
			}
			m->f->extended_data = &m->f->data[0];
			m->f->linesize[0] = out_sz;

			// set custom allocate func
			m->ctx->get_buffer2 = get_frame_buffer;

			int r = avcodec_decode_audio4(m->ctx, m->f, &m->got, &pkt);
			if(r < 0) {
				static char error_buffer[255];
				av_strerror(r, error_buffer, sizeof(error_buffer));
				av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
			}
			if(r > 0 && in_sz > r) {
				av_log(m->ctx, AV_LOG_DEBUG, "return positive result %d\n", r);
			}

			av_log(m->ctx, AV_LOG_DEBUG, "aac_decode, channels layout: %lu, channels: %d, nb_samples: %d, line size: %d\n", m->f->channel_layout, m->ctx->channels, m->f->nb_samples, m->f->linesize[0]);

			return r;
		}

		static int copy_frame_data(aacdec_t *m, uint8_t *data, int sz, int plane_id) {
			memcpy(data, m->f->extended_data[plane_id], sz);

			return 0;
		}
	*/
	"C"
	"errors"
	"unsafe"
)
import "log"

type AACDecoder struct {
	m C.aacdec_t
}

func NewAACDecoder(codec string, header []byte) (m *AACDecoder, err error) {
	m = &AACDecoder{}

	codec_id := 0
	switch codec {
	case "aac":
		codec_id = C.AV_CODEC_ID_AAC
	case "mulaw", "ulaw":
		codec_id = C.AV_CODEC_ID_PCM_MULAW
	case "alaw":
		codec_id = C.AV_CODEC_ID_PCM_ALAW
	case "nellymoser":
		codec_id = C.AV_CODEC_ID_NELLYMOSER
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

	//if m.m.c.capabilities&C.CODEC_CAP_DELAY > 0 {
	log.Printf("Audio decoder capabilities, codec:%s, cap:%+v, CODEC_CAP_DELAY:%+v", codec, m.m.c.capabilities, C.CODEC_CAP_DELAY)
	//}

	if int(r) < 0 {
		err = errors.New("open codec failed")
	}

	return
}

func (m *AACDecoder) Release() {
	C.aacdec_release(&m.m)
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

	size := int(m.m.f.nb_samples) * int(m.m.sample_size)
	sample = make([]byte, size*int(m.m.ctx.channels))
	for i := 0; i < int(m.m.ctx.channels); i++ {
		C.copy_frame_data(&m.m, (*C.uint8_t)(unsafe.Pointer(&sample[i*size])), C.int(size), C.int(i))
	}

	return
}

func (m *AACDecoder) Decode2(data []byte) (sample []byte, err error) {
	// sampleSize := 2
	// if m.m.ctx.sample_fmt == C.AV_SAMPLE_FMT_FLT || m.m.ctx.sample_fmt == AV_SAMPLE_FMT_FLTP {
	// 	sampleSize = 4
	// }
	// size := int(m.m.f.nb_samples) * sampleSize
	sample = make([]byte, 8192*int(m.m.ctx.channels))
	log.Println("channels:", int(m.m.ctx.channels), "sample size:", len(sample))
	r := C.aacdec_decode2(
		&m.m,
		(*C.uint8_t)(unsafe.Pointer(&sample[0])),
		(C.int)(len(sample)),
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

	// sampleSize := 2
	// if m.m.ctx.sample_fmt == C.AV_SAMPLE_FMT_FLT || m.m.ctx.sample_fmt == AV_SAMPLE_FMT_FLTP {
	// 	sampleSize = 4
	// }

	// size := int(int(m.m.f.nb_samples) * sampleSize)
	// sample = make([]byte, size*int(m.m.ctx.channels))
	// for i := 0; i < channels; i++ {
	// 	C.copy_frame_data(&m.m, (*C.uint8_t)(unsafe.Pointer(&sample[i*size])), C.int(size), C.int(i))
	// }

	return
}
