package codec

import (
	/*
					#cgo CFLAGS: -I/usr/local/include
					#cgo LDFLAGS: -L/usr/local/lib  -lavformat -lavcodec -lavresample -lavutil -lx264 -lz -ldl -lm


					#include <stdio.h>
					#include "libavcodec/avcodec.h"
					#include "libavutil/avutil.h"
					#include "libavutil/channel_layout.h"
					#include <string.h>

					typedef struct {
						AVCodec *c;
						AVCodecContext *ctx;
						AVFrame *f;
						int got;
						uint8_t buf[1024*10]; int size;
						int samplerate;
						int bitrate;
						int channels;
						int channel_layout;
						int profile;
						int64_t pts;
						int64_t dts;
						uint8_t* channels_buf[2];
					} aacenc_t ;

					static int select_channel_layout(AVCodec *codec) {
						const uint64_t *p;
						uint64_t best_ch_layout = 0;
						int best_nb_channels   = 0;
						if (!codec->channel_layouts)
							return AV_CH_LAYOUT_STEREO;
						p = codec->channel_layouts;
						while (*p) {
							int nb_channels = av_get_channel_layout_nb_channels(*p);
							if (nb_channels == 2) {
								best_ch_layout    = *p;
								best_nb_channels = nb_channels;
							}
							p++;
						}
						return best_ch_layout;
					}

					static int aacenc_new(aacenc_t *m) {
						m->c = avcodec_find_encoder(AV_CODEC_ID_AAC);

						m->ctx = avcodec_alloc_context3(m->c);
						m->ctx->sample_fmt = AV_SAMPLE_FMT_S16;
						m->ctx->sample_rate = m->samplerate;
						m->ctx->bit_rate = m->bitrate;
						m->ctx->channels = m->channels;
						m->ctx->channel_layout = m->channel_layout;
						m->ctx->profile = m->profile;
						m->ctx->flags |= CODEC_FLAG_GLOBAL_HEADER;
				  		m->ctx->strict_std_compliance = FF_COMPLIANCE_EXPERIMENTAL;

						int r = avcodec_open2(m->ctx, m->c, NULL);
						if(r != 0) {
							static char error_buffer[255];
							av_strerror(r, error_buffer, sizeof(error_buffer));
							av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);
						}

						av_log(m->ctx, AV_LOG_DEBUG, "extra %d\n", m->ctx->extradata_size);
						av_log(m->ctx, AV_LOG_DEBUG, "frame size %d\n", m->ctx->frame_size);
						av_log(m->ctx, AV_LOG_DEBUG, "Audio encoder:, channels: %d, ch_layout: %ld, sample_fmt: %d, planar: %d\n", m->ctx->channels,m->ctx->channel_layout,m->ctx->sample_fmt,av_sample_fmt_is_planar(m->ctx->sample_fmt));

						m->f = av_frame_alloc();
						m->f->nb_samples = m->ctx->frame_size;
						m->f->format = m->ctx->sample_fmt;
						m->f->channel_layout = m->ctx->channel_layout;

						return r;
					}

					static void bind_buf(aacenc_t *m, uint8_t* samples) {
						int ret = avcodec_fill_audio_frame(m->f, 2, AV_SAMPLE_FMT_S16,(const uint8_t*)samples, 4096, 0);

						if (ret < 0) {
							av_log(m->ctx, AV_LOG_DEBUG, "Bind buffer failed ...\n");
							return;
						}
					}

					static void aacenc_encode(aacenc_t *m) {
						AVPacket pkt;
						av_init_packet(&pkt);
						pkt.data = m->buf;
						pkt.size = sizeof(m->buf);
						avcodec_encode_audio2(m->ctx, &pkt, m->f, &m->got);
						av_log(m->ctx, AV_LOG_DEBUG, "got %d size %d, pkt_dts:%ld, frame_dts:%ld\n", m->got, pkt.size,pkt.dts,m->f->pts);

						m->size = pkt.size;
						m->pts = pkt.pts;
						m->dts = pkt.dts;
					}

					static int get_buffer_size(aacenc_t *m) {
						int ret, planar, linesize;

						planar = av_sample_fmt_is_planar(AV_SAMPLE_FMT_S16);
		    			ret =  av_samples_get_buffer_size(&linesize, m->ctx->channels, m->ctx->frame_size,
		                                             AV_SAMPLE_FMT_S16, 0);
						av_log(m->ctx, AV_LOG_DEBUG, "Planar: %d, line size %d\n", planar, linesize);

						return ret;
					}
	*/
	"C"
	"errors"
	"unsafe"
)
import "log"

const (
	FF_PROFILE_AAC_MAIN  = 0
	FF_PROFILE_AAC_LOW   = 1
	FF_PROFILE_AAC_SSR   = 2
	FF_PROFILE_AAC_LTP   = 3
	FF_PROFILE_AAC_HE    = 4
	FF_PROFILE_AAC_HE_V2 = 28
	FF_PROFILE_AAC_LD    = 22
	FF_PROFILE_AAC_ELD   = 38
)

const (
	AV_CH_LAYOUT_RIGHT  = 1
	AV_CH_LAYOUT_LEFT   = 2
	AV_CH_LAYOUT_CENTER = 4
	AV_CH_LAYOUT_STEREO = 3
	AV_CH_LAYOUT_MONO   = 4
)

const (
	AV_SAMPLE_FMT_NONE int32 = -1
	AV_SAMPLE_FMT_U8   int32 = 0 ///< unsigned 8 bits
	AV_SAMPLE_FMT_S16  int32 = 1 ///< signed 16 bits
	AV_SAMPLE_FMT_S32  int32 = 2 ///< signed 32 bits
	AV_SAMPLE_FMT_FLT  int32 = 3 ///< float
	AV_SAMPLE_FMT_DBL  int32 = 4 ///< double

	AV_SAMPLE_FMT_U8P  int32 = 5 ///< unsigned 8 bits, planar
	AV_SAMPLE_FMT_S16P int32 = 6 ///< signed 16 bits, planar
	AV_SAMPLE_FMT_S32P int32 = 7 ///< signed 32 bits, planar
	AV_SAMPLE_FMT_FLTP int32 = 8 ///< float, planar
	AV_SAMPLE_FMT_DBLP int32 = 9 ///< double, planar
)

type AChannelSamples []byte

type AACOut struct {
	Data []byte
	Pts  int64
	Dts  int64
}

type AACEncoder struct {
	m             C.aacenc_t
	SampleRate    int
	BitRate       int
	Channels      int
	ChannelLayout int
	Profile       int
	Header        []byte
	bufL          [8192]byte
	bufR          [8192]byte
	bufLen        int
	sampleSize    int
	frameSize     int
}

// only supported fltp,stereo,44100khz. If you need other config, it's easy to modify code
func NewAACEncoder() (m *AACEncoder, err error) {
	m = &AACEncoder{}
	m.m.samplerate = 44100
	m.m.bitrate = 50000
	m.m.channels = 2

	avLock.Lock()
	defer avLock.Unlock()

	r := C.aacenc_new(&m.m)
	if int(r) != 0 {
		err = errors.New("open codec failed")
		return
	}

	m.Header = make([]byte, (int)(m.m.ctx.extradata_size))
	C.memcpy(
		unsafe.Pointer(&m.Header[0]),
		unsafe.Pointer(&m.m.ctx.extradata),
		(C.size_t)(len(m.Header)),
	)
	return
}

func (m *AACEncoder) Init() error {
	m.m.samplerate = C.int(m.SampleRate)
	m.m.bitrate = C.int(m.BitRate)
	m.m.channels = C.int(m.Channels)
	m.m.channel_layout = C.int(m.ChannelLayout)
	m.m.profile = C.int(m.Profile)

	avLock.Lock()
	defer avLock.Unlock()

	r := C.aacenc_new(&m.m)
	if int(r) != 0 {
		return errors.New("open codec failed")
	}

	// AV_SAMPLE_FMT_S16
	m.sampleSize = 2
	m.frameSize = int(m.m.ctx.frame_size)

	C.bind_buf(&m.m, (*C.uint8_t)(unsafe.Pointer(&m.bufR[0])))

	extraSz := (int)(m.m.ctx.extradata_size)
	if extraSz > 0 {
		m.Header = make([]byte, (int)(m.m.ctx.extradata_size))
		C.memcpy(
			unsafe.Pointer(&m.Header[0]),
			unsafe.Pointer(&m.m.ctx.extradata),
			(C.size_t)(len(m.Header)),
		)
	} else {
		log.Println("AAC codec, extra size: 0")
	}

	return nil
}

func (m *AACEncoder) GetFrameSize() C.int {
	return m.m.ctx.frame_size
}

func (m *AACEncoder) GetBufferSize() C.int {
	return C.get_buffer_size(&m.m)
}

func (m *AACEncoder) GetNbSamples() C.int {
	return m.m.f.nb_samples
}

func (m *AACEncoder) Encode(sample []byte) (ret []byte, err error) {
	m.m.f.data[0] = (*C.uint8_t)(unsafe.Pointer(&sample[0]))
	m.m.f.data[1] = (*C.uint8_t)(unsafe.Pointer(&sample[4096]))

	C.aacenc_encode(&m.m)

	if int(m.m.got) == 0 {
		err = errors.New("no data")
		return
	}

	ret = make([]byte, (int)(m.m.size))
	C.memcpy(
		unsafe.Pointer(&ret[0]),
		unsafe.Pointer(&m.m.buf[0]),
		(C.size_t)(m.m.size),
	)

	return
}

func (m *AACEncoder) samplesToBuf(channels []AChannelSamples, samplesCount, samplesOffset int) int {
	mixedSampleSize := m.Channels * m.sampleSize

	// copy samples count
	sz := samplesCount - samplesOffset
	if m.frameSize-m.bufLen < sz {
		sz = m.frameSize - m.bufLen
	}

	// bufLen -> buffer size in samples
	for i := 0; i < sz; i++ {
		sampleBuf := m.bufR[(m.bufLen+i)*mixedSampleSize:]
		for j := 0; j < m.Channels; j++ {
			for k := 0; k < m.sampleSize; k++ {
				sampleBuf[j*m.sampleSize+k] = channels[j][(samplesOffset+i)*m.sampleSize+k]
			}
		}
	}
	m.bufLen += sz

	// buf is full, return remind samples count
	return sz
}

func (m *AACEncoder) Encode2(channels []AChannelSamples) (ret *AACOut, err error) {
	if len(channels) != m.Channels {
		err = errors.New("Channels number not equal")
		return
	}

	// channel buf size
	channelBufSz := 0
	for _, v := range channels {
		if channelBufSz == 0 {
			channelBufSz = len(v)
			continue
		}

		if channelBufSz != len(v) {
			err = errors.New("channels size not equal")
			return
		}
	}

	samplesCount := channelBufSz / m.sampleSize

	// copy samples to codec buf
	n := m.samplesToBuf(channels, samplesCount, 0)

	// until codec buf not full, return no_data
	if m.bufLen < m.frameSize {
		err = errors.New("more data")
		return
	}

	// encode samples
	C.aacenc_encode(&m.m)

	// reset bufLen
	m.bufLen = 0

	// after encode, copy remind samples
	if n < samplesCount {
		//log.Println("Copy remaining data:", samplesCount-n)
		m.samplesToBuf(channels, samplesCount, n)
	}

	// check got flag
	if int(m.m.got) == 0 {
		err = errors.New("no data")
		return
	}

	//log.Println("Encoded data size:", m.m.size)

	// extract and return encoded data
	ret = &AACOut{
		Data: make([]byte, (int)(m.m.size)),
		Pts:  int64(m.m.pts),
		Dts:  int64(m.m.dts),
	}

	C.memcpy(
		unsafe.Pointer(&ret.Data[0]),
		unsafe.Pointer(&m.m.buf[0]),
		(C.size_t)(m.m.size),
	)

	return
}
