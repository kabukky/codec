package codec

import (
	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -L/usr/local/lib  -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm

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
			int got;
			AVCodec *c;
			AVCodecContext *ctx;
			AVFrame *f;
			AVPacket pkt;
		} h264enc_t;

		typedef struct {
			AVStream *video_st;
			AVFormatContext *ctx;
			AVOutputFormat *fmt;
			char *filename;
		} avformat_t;

		static int avformat_new(avformat_t *m, char *filename) {
			m->filename = filename;

			m->ctx = avformat_alloc_context();

			m->fmt = av_guess_format(NULL, filename, NULL);
			if (!m->fmt) {
				printf("Could not deduce output format from file extension: using MPEG.\n");
				m->fmt = av_guess_format("mpeg", NULL, NULL);
			}

			m->ctx->oformat = m->fmt;

			// Open output file
			if (avio_open(&m->ctx->pb, filename, AVIO_FLAG_WRITE) < 0) {
				fprintf(stderr, "Could not open '%s'\n", filename);

				return 1;
			}

			return 0;
		}

		static int add_video_stream(avformat_t *m, h264enc_t *enc) {
			m->video_st = avformat_new_stream(m->ctx, enc->c);
			printf("%s\n",m->video_st->codec->codec->long_name);

			m->video_st->codec->width  		= enc->ctx->width;
			m->video_st->codec->height 		= enc->ctx->height;
			m->video_st->codec->bit_rate 	= enc->ctx->bit_rate;
			m->video_st->codec->time_base 	= enc->ctx->time_base;

			m->video_st->time_base	= enc->ctx->time_base;
			m->video_st->codec->gop_size	= enc->ctx->gop_size;
			m->video_st->codec->pix_fmt 	= enc->ctx->pix_fmt;
			//m->ctx->flags |= CODEC_FLAG_GLOBAL_HEADER;

			m->video_st->codec->flags |= CODEC_FLAG_GLOBAL_HEADER;

			if (avcodec_open2(m->video_st->codec, NULL, NULL) < 0) {
				fprintf(stderr, "could not open codec\n");
			}

			av_dump_format(m->ctx, 0, m->filename, 1);

			return 0;
		}

		static int write_header(avformat_t *m) {
			// Write the stream header, if any.
			avformat_write_header(m->ctx, NULL);

			return 0;
		}

		static int write_pkt(avformat_t *m, AVPacket *pkt) {
			//printf("pkt pts %ld, %d %d\n",pkt->pts, m->video_st->codec->time_base.num,m->video_st->codec->time_base.den);
			//printf("pkt pts %ld, %d %d\n",pkt->pts, m->video_st->time_base.num,m->video_st->time_base.den);

			av_packet_rescale_ts(pkt, m->video_st->codec->time_base, m->video_st->time_base);
			int64_t tt = 12345678;
			//printf("pkt pts %ld, t %ld\n",pkt->pts,tt);
			pkt->stream_index = m->video_st->index;
			int ret = av_interleaved_write_frame(m->ctx, pkt);
			return ret;
		}

		static int complete(avformat_t *m) {
			av_write_trailer(m->ctx);

			//close_stream(m->ctx, m->video_st);
			avio_close(m->ctx->pb);
		}
	*/
	"C"
	//	"errors"
	//	"image"
	//	"strings"
	"unsafe"
	//"log"
)

type AVFormat struct {
	m      C.avformat_t
	fname  string
	Header []byte
}

func CreateAVFormat(fname string) *AVFormat {
	f := &AVFormat{}

	f.fname = fname
	C.avformat_new(&f.m, C.CString(fname))

	return f
}

func (f *AVFormat) AddVideoStream(enc *H264Encoder) {
	C.add_video_stream(&f.m, &enc.m)
	f.Header = fromCPtr(unsafe.Pointer(f.m.video_st.codec.extradata), (int)(f.m.video_st.codec.extradata_size))
}

func (f *AVFormat) WriteHeader() {
	C.write_header(&f.m)
}

func (f *AVFormat) WritePacket(enc *H264Encoder) {
	C.write_pkt(&f.m, &enc.m.pkt)
}

func (f *AVFormat) Close() {
	C.complete(&f.m)
}
