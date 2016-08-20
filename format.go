package codec

import (
	/*
		#cgo CFLAGS: -I/usr/local/include
		#cgo LDFLAGS: -L/usr/local/lib  -lavformat -lavcodec -lavresample -lavutil -lfdk-aac -lx264 -lz -ldl -lm

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

		typedef struct {
			AVStream *video_st;
			AVStream *audio_st;
			AVFormatContext *ctx;
			AVOutputFormat *fmt;
			char *filename;
			AVPacket pkt;
			AVCodec *c;
		} avformat_t;

		static int avformat_new(avformat_t *m, char *filename) {
			m->filename = filename;

			m->ctx = avformat_alloc_context();

			m->fmt = av_guess_format(NULL, filename, NULL);
			if (!m->fmt) {
				av_log(m->ctx, AV_LOG_DEBUG, "Could not deduce output format from file extension: using MPEG.\n");
				m->fmt = av_guess_format("mpeg", NULL, NULL);
			}

			m->ctx->oformat = m->fmt;

			// Open output file
			if (avio_open(&m->ctx->pb, filename, AVIO_FLAG_WRITE) < 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "Could not open '%s'\n", filename);

				return -1;
			}
			av_init_packet(&m->pkt);

			return 0;
		}

		static int add_video_stream(avformat_t *m, h264enc_t *enc) {
			m->video_st = avformat_new_stream(m->ctx, enc->c);
			//printf("%s\n",m->video_st->codec->codec->long_name);

			m->video_st->codec->width  		= enc->ctx->width;
			m->video_st->codec->height 		= enc->ctx->height;
			m->video_st->codec->bit_rate 	= enc->ctx->bit_rate;
			m->video_st->codec->time_base 	= enc->ctx->time_base;

			m->video_st->time_base	= enc->ctx->time_base;
			m->video_st->codec->gop_size	= enc->ctx->gop_size;
			m->video_st->codec->pix_fmt 	= enc->ctx->pix_fmt;
			m->video_st->codec->flags |= CODEC_FLAG_GLOBAL_HEADER;

			if (avcodec_open2(m->video_st->codec, NULL, NULL) < 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "could not open codec\n");
			}

			av_dump_format(m->ctx, 0, m->filename, 1);

			return 0;
		}

		static int add_video_stream2(avformat_t *m) {
			m->video_st = avformat_new_stream(m->ctx, NULL);

			return 0;
		}

		static int add_audio_stream2(avformat_t *m) {
			m->c = avcodec_find_encoder(AV_CODEC_ID_AAC);
			m->audio_st = avformat_new_stream(m->ctx, m->c);

			return 0;
		}

		static int open_video_stream2(avformat_t *m) {
			if (avcodec_open2(m->video_st->codec, NULL, NULL) < 0) {
				fprintf(stderr, "could not open codec\n");
				av_log(m->ctx, AV_LOG_DEBUG, "could not open codec\n");
			}

			//av_dump_format(m->ctx, 0, m->filename, 1);

			return 0;

		}

		static int open_codec(avformat_t *m) {
			int r = avcodec_open2(m->audio_st->codec, NULL, NULL);
			if (r < 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "could not open codec\n");
				return r;
			}

			av_dump_format(m->ctx, 0, m->filename, 1);

			return r;
		}

		static int write_header(avformat_t *m) {
			av_dump_format(m->ctx, 0, m->filename, 1);

			// Write the stream header, if any.
			avformat_write_header(m->ctx, NULL);

			return 0;
		}

		static int write_pkt(avformat_t *m, AVPacket *pkt) {
			//printf("pkt pts %ld, %d %d\n",pkt->pts, m->video_st->codec->time_base.num,m->video_st->codec->time_base.den);
			//printf("pkt pts %ld, %d %d\n",pkt->pts, m->video_st->time_base.num,m->video_st->time_base.den);

			av_packet_rescale_ts(pkt, m->video_st->codec->time_base, m->video_st->time_base);
			av_log(m->ctx, AV_LOG_DEBUG, "write packet: pts: %ld, dts: %ld\n", pkt->pts, pkt->dts);

			int64_t tt = 12345678;
			//printf("pkt pts %ld, t %ld\n",pkt->pts,tt);
			pkt->stream_index = m->video_st->index;
			int ret = av_interleaved_write_frame(m->ctx, pkt);
			//int ret = av_write_frame(m->ctx, pkt);
			//int ret = 0;

			// static char error_buffer[255];
			// av_strerror(ret, error_buffer, sizeof(error_buffer));
			// av_log(m->ctx, AV_LOG_DEBUG, "write packet: pts: %ld, dts: %ld, error: %s\n", pkt->pts, pkt->dts, error_buffer);

			return ret;
		}

		static int write_pkt2(avformat_t *m, uint8_t *data, int len, int64_t tm, int isKeyFrame) {
			AVPacket pkt;
			av_init_packet(&pkt);
			pkt.data = data;
			pkt.size = len;
			pkt.stream_index = m->video_st->index;
			pkt.pts = tm;
			pkt.dts = tm;
			printf("pts: %ld, tm: %ld\n", pkt.pts, tm);
			av_packet_rescale_ts(&pkt, m->video_st->codec->time_base, m->video_st->time_base);
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, len: %d\n", pkt.pts, pkt.dts,len);
			if(isKeyFrame) {
				pkt.flags |= AV_PKT_FLAG_KEY;
			}
			//int ret = av_interleaved_write_frame(m->ctx, &pkt);
			int ret = av_write_frame(m->ctx, &pkt);

			static char error_buffer[255];
			av_strerror(ret, error_buffer, sizeof(error_buffer));
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, error: %s\n", pkt.pts, tm, error_buffer);

			return ret;
		}

		static int write_audio_pkt2(avformat_t *m, uint8_t *data, int len, int64_t tm) {
			AVPacket pkt;
			av_init_packet(&pkt);
			pkt.data = data;
			pkt.size = len;
			pkt.stream_index = m->audio_st->index;
			pkt.pts = tm;
			pkt.dts = tm;

			av_packet_rescale_ts(&pkt, m->audio_st->codec->time_base, m->audio_st->time_base);
			av_log(m->ctx, AV_LOG_DEBUG, "rescale: pts: %ld, dts: %ld, len: %d\n", pkt.pts, pkt.dts,len);

			int ret = av_interleaved_write_frame(m->ctx, &pkt);

			static char error_buffer[255];
			av_strerror(ret, error_buffer, sizeof(error_buffer));
			av_log(m->ctx, AV_LOG_DEBUG, "write: pts: %ld, dts: %ld, error: %s\n", pkt.pts, tm, error_buffer);

			return ret;
		}

		static int complete(avformat_t *m) {
			av_write_trailer(m->ctx);

			//close_stream(m->ctx, m->video_st);
			avio_close(m->ctx->pb);
		}
	*/
	"C"
	"errors"
	"image"
	//	"strings"
	"bytes"
	//"log"
	"unsafe"
)
import "fmt"

const (
	AVMEDIA_TYPE_VIDEO = 0
	AVMEDIA_TYPE_AUDIO = 1
)

type AVRational struct {
	Num, Den int
}

type AVStreamInfo struct {
	// video
	W        int
	H        int
	Pixfmt   image.YCbCrSubsampleRatio
	Bitrate  int
	TimeBase AVRational
	GopSize  int

	// audio
	SampleRate    int
	ABitRate      int
	Channels      int
	ChannelLayout int
	Profile       int
	SampleFmt     int32
}

type AVFormat struct {
	m           C.avformat_t
	videoStream AVStreamInfo
	fname       string
	Header      []byte
	frameBuf    bytes.Buffer
	pts         int64
}

func CreateAVFormat(fname string) (*AVFormat, error) {
	f := &AVFormat{}

	f.fname = fname
	r := C.avformat_new(&f.m, C.CString(fname))
	if int(r) < 0 {
		err := errors.New("Create format failed")
		return nil, err
	}
	return f, nil
}

func (f *AVFormat) AddVideoStream(enc *H264Encoder) {
	C.add_video_stream(&f.m, &enc.m)
	f.Header = fromCPtr(unsafe.Pointer(f.m.video_st.codec.extradata), (int)(f.m.video_st.codec.extradata_size))
}

func (f *AVFormat) AddVideoStream2(info *AVStreamInfo, extra []byte) (err error) {
	// add video stream
	C.add_video_stream2(&f.m)

	// setup codec
	f.m.video_st.codec.codec_type = C.AVMEDIA_TYPE_VIDEO
	f.m.video_st.codec.codec_id = C.AV_CODEC_ID_H264

	if extra != nil {
		f.m.video_st.codec.extradata = (*C.uint8_t)(unsafe.Pointer(&extra[0]))
		f.m.video_st.codec.extradata_size = (C.int)(len(extra))
	}

	f.m.video_st.codec.width = C.int(info.W)
	f.m.video_st.codec.height = C.int(info.H)
	f.m.video_st.codec.bit_rate = C.int(info.Bitrate)
	f.m.video_st.codec.time_base.num = C.int(info.TimeBase.Num)
	f.m.video_st.codec.time_base.den = C.int(info.TimeBase.Den)
	f.m.video_st.codec.gop_size = C.int(info.GopSize)
	// pix_fmt has int32 type, but not C.int32() method
	switch info.Pixfmt {
	case image.YCbCrSubsampleRatio444:
		f.m.video_st.codec.pix_fmt = C.PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		f.m.video_st.codec.pix_fmt = C.PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		f.m.video_st.codec.pix_fmt = C.PIX_FMT_YUV420P
	}
	f.m.video_st.codec.flags |= C.CODEC_FLAG_GLOBAL_HEADER

	// setup stream
	f.m.video_st.time_base.num = 1
	f.m.video_st.time_base.den = 1000

	return
}

func (f *AVFormat) AddAudioStream2(info *AVStreamInfo, extra []byte) (err error) {
	// add video stream
	C.add_audio_stream2(&f.m)

	// setup codec
	f.m.audio_st.codec.codec_type = C.AVMEDIA_TYPE_AUDIO
	f.m.audio_st.codec.codec_id = C.AV_CODEC_ID_AAC

	f.m.audio_st.codec.sample_rate = C.int(info.SampleRate)
	f.m.audio_st.codec.channels = C.int(info.Channels)
	f.m.audio_st.codec.channel_layout = C.uint64_t(info.ChannelLayout)
	f.m.audio_st.codec.profile = C.int(info.Profile)
	//f.m.audio_st.codec.sample_fmt = info.SampleFmt
	f.m.audio_st.codec.sample_fmt = C.AV_SAMPLE_FMT_S16

	if extra != nil {
		f.m.audio_st.codec.extradata = (*C.uint8_t)(unsafe.Pointer(&extra))
		f.m.audio_st.codec.extradata_size = (C.int)(len(extra))
	}

	f.m.audio_st.codec.flags |= C.CODEC_FLAG_GLOBAL_HEADER

	// setup stream
	f.m.audio_st.time_base.num = C.int(1)
	f.m.audio_st.time_base.den = C.int(info.SampleRate)

	avLock.Lock()
	defer avLock.Unlock()

	r := C.open_codec(&f.m)
	if int(r) != 0 {
		return errors.New("Failed open AAC format codec ...")
	}

	return
}

func (f *AVFormat) AddAudioPcmStream() (err error) {
	// add video stream
	C.add_audio_stream2(&f.m)

	// setup codec
	f.m.audio_st.codec.codec_type = C.AVMEDIA_TYPE_AUDIO
	f.m.audio_st.codec.codec_id = C.AV_CODEC_ID_PCM_S16LE

	f.m.audio_st.codec.sample_rate = 8000
	f.m.audio_st.codec.channels = 1
	f.m.audio_st.codec.channel_layout = 4
	f.m.audio_st.codec.sample_fmt = C.AV_SAMPLE_FMT_S16

	// setup stream
	f.m.audio_st.time_base.num = C.int(1)
	f.m.audio_st.time_base.den = C.int(8000)

	return
}

func (f *AVFormat) WriteHeader() {
	C.write_header(&f.m)
}

func (f *AVFormat) WritePacket(enc *H264Encoder) {
	C.write_pkt(&f.m, &enc.m.pkt)
}

func (f *AVFormat) WritePacket2(o *H264Out) {
	defer o.Free()

	C.write_pkt(&f.m, &o.pkt)
}

func (f *AVFormat) WriteVideoData(nal []byte, timeStamp uint32, isKeyFrame bool) (err error) {
	//tm := int64(timeStamp)
	ikf := 0
	if isKeyFrame {
		ikf = 1
	}
	r := C.write_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(len(nal)), C.int64_t(f.pts), C.int(ikf))
	f.pts++

	if int(r) != 0 {
		err = errors.New(fmt.Sprintf("Write video data failed, code:%v", r))
		return
	}

	return
}

func (f *AVFormat) WriteAudioData(nal []byte, timeStamp uint32) (err error) {
	r := C.write_audio_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(len(nal)), C.int64_t(f.pts))
	f.pts += 1024

	if int(r) != 0 {
		err = errors.New(fmt.Sprintf("Write audio data failed, code:%v", r))
		return
	}

	return
}

func (f *AVFormat) Close() {
	C.complete(&f.m)
}
