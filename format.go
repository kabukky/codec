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
		#include "libavformat/avformat.h"

		typedef struct {
			AVStream *video_st;
			AVStream *audio_st;
			AVFormatContext *ctx;
			AVBitStreamFilterContext *bsfc;
			AVOutputFormat *fmt;
			char *filename;
			AVPacket pkt;
			AVCodec *c;
			AVCodec *cv;
			int useToAnnexbFilter;
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

			// create bit filter context
			if (m->useToAnnexbFilter) {
				m->bsfc = av_bitstream_filter_init("h264_mp4toannexb");
				if(m->bsfc == NULL) {
					av_log(m->ctx, AV_LOG_DEBUG, "Create bitstream filter h264_mp4toannexb failed\n");

					return -1;
				}
			}

			return 0;
		}

		static int avformat_switch_outfile(avformat_t *m, char *filename) {
			// close output file
			av_write_trailer(m->ctx);
			avio_close(m->ctx->pb);

			// Open output file
			if (avio_open(&m->ctx->pb, filename, AVIO_FLAG_WRITE) < 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "Could not open '%s'\n", filename);

				return -1;
			}

			// Write the stream header, if any.
			avformat_write_header(m->ctx, NULL);

			return 0;
		}

		static int add_video_stream2(avformat_t *m) {
			m->cv = avcodec_find_encoder(AV_CODEC_ID_H264);
			m->video_st = avformat_new_stream(m->ctx, m->cv);

			return 0;
		}

		static int add_audio_stream2(avformat_t *m) {
			m->c = avcodec_find_encoder(AV_CODEC_ID_AAC);
			m->audio_st = avformat_new_stream(m->ctx, m->c);

			return 0;
		}

		static int open_video_stream2(avformat_t *m) {
			int i;

			AVDictionary *codec_options = NULL;
			av_dict_set( &codec_options, "preset", "veryfast", 0 );
			av_dict_set( &codec_options, "profile", "high", 0 );
			av_dict_set( &codec_options, "level", "41", 0 );

			if (avcodec_open2(m->video_st->codec, NULL, &codec_options) < 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "could not open codec\n");
			}

			for(i=0;i<m->video_st->codec->extradata_size;++i) {
				av_log(m->ctx, AV_LOG_DEBUG, "0x%02x ", m->video_st->codec->extradata[i]);
			}

			//av_dump_format(m->ctx, 0, m->filename, 1);
			av_log(m->ctx, AV_LOG_DEBUG, "AVFormat, video encoder created");

			return 0;

		}

		static void set_video_extradata(avformat_t *m, uint8_t *extra, int size) {
			if (av_reallocp(&m->video_st->codec->extradata, size) != 0) {
				av_log(m->ctx, AV_LOG_DEBUG, "allocate memory for extradata failed\n");

				return;
			}

			memcpy(m->video_st->codec->extradata, extra, size);
			m->video_st->codec->extradata_size = size;
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
			if (data == NULL) {
				return 0;
			}

			m->pkt.stream_index = m->video_st->index;
			m->pkt.pts = tm;
			m->pkt.dts = tm;
			if(isKeyFrame) {
				m->pkt.flags |= AV_PKT_FLAG_KEY;
			}

			// set pkt data
			if (m->useToAnnexbFilter) {
				int ret = av_bitstream_filter_filter(m->bsfc, m->video_st->codec, NULL,
													&m->pkt.data, &m->pkt.size,
													data, len,
													m->pkt.flags & AV_PKT_FLAG_KEY);

				if (ret > 0) {
					// non-zero positive, you have new memory allocated,
					// keep it referenced in the AVBuffer
					m->pkt.buf = av_buffer_create(m->pkt.data, m->pkt.size, av_buffer_default_free, NULL, 0);
				} else if (ret < 0) {
					// handle failure here
					static char error_buffer[255];
					av_strerror(ret, error_buffer, sizeof(error_buffer));
					av_log(m->ctx, AV_LOG_DEBUG, "write_pkt2: apply filter failed, err: %s\n", error_buffer);
				}
			} else {
				m->pkt.data = data;
				m->pkt.size = len;
			}

			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, tm: %ld\n", m->pkt.pts, tm);
			av_packet_rescale_ts(&m->pkt, m->video_st->codec->time_base, m->video_st->time_base);
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, len: %d\n", m->pkt.pts, m->pkt.dts,len);

			//int ret = av_interleaved_write_frame(m->ctx, &m->pkt);
			int ret = av_write_frame(m->ctx, &m->pkt);

			static char error_buffer[255];
			av_strerror(ret, error_buffer, sizeof(error_buffer));
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, error: %s\n", m->pkt.pts, tm, error_buffer);

			return ret;
		}

		static int write_pkt3(avformat_t *m, AVPacket *pkt, uint8_t *data) {
			int i =0;

			av_log(m->ctx, AV_LOG_DEBUG, "m: %p, pkt: %p, data: %p\n", m, pkt, data);

			if (data != NULL) {
				pkt->data = data;
			}

			m->pkt.stream_index = m->video_st->index;
			m->pkt.pts = pkt->pts;
			m->pkt.dts = pkt->dts;
			m->pkt.flags = pkt->flags;

			// set pkt data
			if (m->useToAnnexbFilter) {
				for(i=0;i<m->video_st->codec->extradata_size;++i) {
					av_log(m->ctx, AV_LOG_DEBUG, "0x%02x ", m->video_st->codec->extradata[i]);
				}
				av_log(m->ctx, AV_LOG_DEBUG, "\nflags: %02x, isKey: %d\n",pkt->flags,pkt->flags & AV_PKT_FLAG_KEY);

				int ret = av_bitstream_filter_filter(m->bsfc, m->video_st->codec, NULL,
													&m->pkt.data, &m->pkt.size,
													pkt->data, pkt->size,
													pkt->flags & AV_PKT_FLAG_KEY);

				if (ret > 0) {
					// non-zero positive, you have new memory allocated,
					// keep it referenced in the AVBuffer
					m->pkt.buf = av_buffer_create(m->pkt.data, m->pkt.size, av_buffer_default_free, NULL, 0);
				} else if (ret < 0) {
					// handle failure here
					static char error_buffer[255];
					av_strerror(ret, error_buffer, sizeof(error_buffer));
					av_log(m->ctx, AV_LOG_DEBUG, "write_pkt3: apply filter failed, err: %s\n", error_buffer);
				}
			} else {
				av_log(m->ctx, AV_LOG_DEBUG, "write_pkt3, pkt->buffer: %p\n", pkt->buf);

				m->pkt.data = pkt->data;
				m->pkt.size = pkt->size;
			}

			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, tm: %ld, codec time base: %d/%d, stream time base: %d/%d\n", m->pkt.pts, pkt->pts,
				m->video_st->codec->time_base.num, m->video_st->codec->time_base.den, m->video_st->time_base.num, m->video_st->time_base.den);
			av_packet_rescale_ts(&m->pkt, m->video_st->codec->time_base, m->video_st->time_base);
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, len: %d\n", m->pkt.pts, m->pkt.dts, m->pkt.size);

			int ret = av_interleaved_write_frame(m->ctx, &m->pkt);
			//int ret = av_interleaved_write_frame(m->ctx, pkt);

			//int ret = av_write_frame(m->ctx, &m->pkt);

			static char error_buffer[255];
			av_strerror(ret, error_buffer, sizeof(error_buffer));
			av_log(m->ctx, AV_LOG_DEBUG, "pts: %ld, dts: %ld, error: %s\n", m->pkt.pts, pkt->pts, error_buffer);

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
			av_log(m->ctx, AV_LOG_INFO, "rescale: pts: %ld, dts: %ld, len: %d\n", pkt.pts, pkt.dts,len);

			int ret = av_interleaved_write_frame(m->ctx, &pkt);

			static char error_buffer[255];
			av_strerror(ret, error_buffer, sizeof(error_buffer));
			av_log(m->ctx, AV_LOG_INFO, "write: pts: %ld, dts: %ld, error: %s\n", pkt.pts, tm, error_buffer);

			return ret;
		}

		static int complete(avformat_t *m) {
			int ret;
			static char error_buffer[255];

			av_log(m->ctx, AV_LOG_DEBUG, "Release AVFormats\n");

			// ret = av_interleaved_write_frame(m->ctx, NULL);
			// if (ret < 0) {
			// 	av_strerror(ret, error_buffer, sizeof(error_buffer));
			// 	av_log(m->ctx, AV_LOG_DEBUG, "av_interleaved_write_frame, error: %s\n",  error_buffer);
			// }

			av_write_trailer(m->ctx);

			avcodec_close(m->video_st->codec);
			avcodec_close( m->audio_st->codec);
			avio_close(m->ctx->pb);
			avformat_free_context(m->ctx);
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
import (
	"fmt"
	"log"
)

const (
	AVMEDIA_TYPE_VIDEO = 0
	AVMEDIA_TYPE_AUDIO = 1
)

type AVRational struct {
	Num, Den int
}

type AVStreamInfo struct {
	// video
	W              int
	H              int
	Pixfmt         image.YCbCrSubsampleRatio
	Bitrate        int
	TimeBase       AVRational
	GopSize        int
	UseMp4ToAnnexb bool
	Tbn            int
	CodecID        int

	// audio
	DisableAudio  bool
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
	vExtra      []byte
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

func CreateAVFormat2(fname string, useToAnnexbFilter bool) (*AVFormat, error) {
	f := &AVFormat{}

	f.fname = fname
	if useToAnnexbFilter {
		f.m.useToAnnexbFilter = 1
	} else {
		f.m.useToAnnexbFilter = 0
	}

	r := C.avformat_new(&f.m, C.CString(fname))
	if int(r) < 0 {
		err := errors.New("Create format failed")
		return nil, err
	}
	return f, nil
}

func (f *AVFormat) APts() *int64 {
	return &f.pts
}

func (f *AVFormat) UseGlobalHeaders() bool {
	c := int(f.m.ctx.oformat.flags) & int(C.AVFMT_GLOBALHEADER)

	return c != 0
}

func (f *AVFormat) SwitchOutFile(fname string) error {
	r := C.avformat_switch_outfile(&f.m, C.CString(fname))
	if int(r) < 0 {
		err := errors.New("Create format failed")
		return err
	}

	return nil
}

func (f *AVFormat) GetVideoEncoder() *H264Encoder {
	return NewH264EncoderFromCtx(f.m.video_st.codec)
}

func (f *AVFormat) GetAacEncoder() *AACEncoder {
	return NewAACEncoderFromCtx(f.m.audio_st.codec)
}

func (f *AVFormat) AddVideoStream2(info *AVStreamInfo, extra []byte) (err error) {
	// add video stream
	C.add_video_stream2(&f.m)

	// setup codec
	f.m.video_st.codec.codec_type = C.AVMEDIA_TYPE_VIDEO

	f.m.video_st.codec.codec_id = C.AV_CODEC_ID_H264
	if info.CodecID > 0 {
		switch info.CodecID {
		case int(C.AV_CODEC_ID_VP8):
			f.m.video_st.codec.codec_id = C.AV_CODEC_ID_VP8
		}
	}

	f.m.video_st.codec.width = C.int(info.W)
	f.m.video_st.codec.height = C.int(info.H)
	f.m.video_st.codec.bit_rate = C.int(info.Bitrate)
	f.m.video_st.codec.time_base.num = C.int(info.TimeBase.Num)
	f.m.video_st.codec.time_base.den = C.int(info.TimeBase.Den)
	f.m.video_st.codec.gop_size = C.int(info.GopSize)

	//f.m.video_st.time_base = f.m.video_st.codec.time_base

	switch info.Pixfmt {
	case image.YCbCrSubsampleRatio444:
		f.m.video_st.codec.pix_fmt = C.AV_PIX_FMT_YUV444P
	case image.YCbCrSubsampleRatio422:
		f.m.video_st.codec.pix_fmt = C.AV_PIX_FMT_YUV422P
	case image.YCbCrSubsampleRatio420:
		f.m.video_st.codec.pix_fmt = C.AV_PIX_FMT_YUV420P
	}

	// h264_mp4toannexb filter use av_free method for extradata
	// we need allocate memory by libav
	// if extra != nil {
	// 	C.set_video_extradata(&f.m, (*C.uint8_t)(unsafe.Pointer(&extra[0])), (C.int)(len(extra)))

	// 	f.vExtra = make([]byte, len(extra))
	// 	copy(f.vExtra, extra)
	// }

	if (int(f.m.ctx.oformat.flags) & int(C.AVFMT_GLOBALHEADER)) != 0 {
		f.m.video_st.codec.flags |= C.AV_CODEC_FLAG_GLOBAL_HEADER
	}

	// setup stream
	f.m.video_st.time_base.num = 1
	f.m.video_st.time_base.den = 1000
	if info.Tbn > 0 {
		f.m.video_st.time_base.den = C.int(info.Tbn)
	}

	avLock.Lock()
	defer avLock.Unlock()

	C.open_video_stream2(&f.m)

	return
}

func (f *AVFormat) AddAudioStream2(info *AVStreamInfo, extra []byte) (err error) {
	// add video stream
	C.add_audio_stream2(&f.m)

	// setup codec
	f.m.audio_st.codec.codec_type = C.AVMEDIA_TYPE_AUDIO
	f.m.audio_st.codec.codec_id = C.AV_CODEC_ID_AAC

	f.m.audio_st.codec.sample_rate = C.int(info.SampleRate)
	f.m.audio_st.codec.bit_rate = C.int(info.ABitRate)
	f.m.audio_st.codec.channels = C.int(info.Channels)
	f.m.audio_st.codec.channel_layout = C.uint64_t(info.ChannelLayout)
	f.m.audio_st.codec.profile = C.int(info.Profile)
	//f.m.audio_st.codec.sample_fmt = info.SampleFmt
	f.m.audio_st.codec.sample_fmt = C.AV_SAMPLE_FMT_S16

	if extra != nil {
		f.m.audio_st.codec.extradata = (*C.uint8_t)(unsafe.Pointer(&extra))
		f.m.audio_st.codec.extradata_size = (C.int)(len(extra))
	}

	if (int(f.m.ctx.oformat.flags) & int(C.AVFMT_GLOBALHEADER)) != 0 {
		f.m.audio_st.codec.flags |= C.AV_CODEC_FLAG_GLOBAL_HEADER
	}

	if f.m.c.capabilities&C.AV_CODEC_CAP_DELAY > 0 {
		log.Printf("Audio codec has delay, cap:%+v, flag:%+v", f.m.c.capabilities, C.AV_CODEC_CAP_DELAY)
	}

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

	//C.write_pkt(&f.m, &o.pkt)
	//o.pkt.data = (*C.uint8_t)(unsafe.Pointer(&o.Data[0]))
	if o.Data != nil && len(o.Data) > 0 {
		C.write_pkt3(&f.m, &o.pkt, (*C.uint8_t)(unsafe.Pointer(&o.Data[0])))
	} else {
		C.write_pkt3(&f.m, &o.pkt, (*C.uint8_t)(unsafe.Pointer(nil)))
	}
}

func (f *AVFormat) PacketVideoData(nal []byte, pts uint32, isKeyFrame bool) *H264Out {
	out := NewH264Out()

	out.Key = isKeyFrame
	out.Data = nal
	// out.Data = make([]byte, len(nal))
	// copy(out.Data, nal)
	//out.pkt.data = (*C.uint8_t)(unsafe.Pointer(&nal[0]))
	out.pkt.size = C.int(len(nal))
	out.pkt.pts = C.int64_t(pts)
	out.pkt.dts = out.pkt.pts

	if isKeyFrame {
		out.pkt.flags |= C.AV_PKT_FLAG_KEY
	}

	return out
}

func (f *AVFormat) WriteVideoData(nal []byte, timeStamp uint32, isKeyFrame bool) (err error) {
	//tm := int64(timeStamp)
	ikf := 0
	if isKeyFrame {
		ikf = 1
	}

	sz := len(nal)
	//	r := C.write_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(sz), C.int64_t(f.pts), C.int(ikf))
	r := C.write_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(sz), C.int64_t(f.pts), C.int(ikf))
	f.pts++

	if int(r) != 0 {
		err = errors.New(fmt.Sprintf("Write video data failed, code:%v", r))
		return
	}

	return
}

func (f *AVFormat) WriteAudioData(nal []byte, timeStamp int64) (err error) {
	//r := C.write_audio_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(len(nal)), C.int64_t(f.pts))
	r := C.write_audio_pkt2(&f.m, (*C.uint8_t)(unsafe.Pointer(&nal[0])), (C.int)(len(nal)), C.int64_t(timeStamp))
	//f.pts = timeStamp

	if int(r) != 0 {
		err = errors.New(fmt.Sprintf("Write audio data failed, code:%v", r))
		return
	}

	return
}

func (f *AVFormat) Close() {
	C.complete(&f.m)
}
