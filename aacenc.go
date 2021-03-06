package codec

import (
	/*
		#cgo linux,amd64 pkg-config: libav_linux_amd64.pc


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
						int release_ctx;
						int release_frame;
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
				  		m->ctx->strict_std_compliance = FF_COMPLIANCE_EXPERIMENTAL;
						m->ctx->flags |= AV_CODEC_FLAG_GLOBAL_HEADER;

						int r = avcodec_open2(m->ctx, m->c, NULL);
						if(r != 0) {
							static char error_buffer[255];
							av_strerror(r, error_buffer, sizeof(error_buffer));
							av_log(m->ctx, AV_LOG_DEBUG, "error %s\n", error_buffer);

							return r;
						}

						m->release_ctx = 1;

						av_log(m->ctx, AV_LOG_DEBUG, "extra %d\n", m->ctx->extradata_size);
						av_log(m->ctx, AV_LOG_DEBUG, "frame size %d\n", m->ctx->frame_size);
						av_log(m->ctx, AV_LOG_DEBUG, "Audio encoder:, channels: %d, ch_layout: %ld, sample_fmt: %d, planar: %d\n", m->ctx->channels,m->ctx->channel_layout,m->ctx->sample_fmt,av_sample_fmt_is_planar(m->ctx->sample_fmt));

						m->f = av_frame_alloc();
						m->f->nb_samples = m->ctx->frame_size;
						m->f->format = m->ctx->sample_fmt;
						m->f->channel_layout = m->ctx->channel_layout;
						m->f->pts =0;

						m->release_frame = 1;

						return r;
					}

					static void aacenc_new_frame(aacenc_t *m) {
						m->f = av_frame_alloc();

						m->f->nb_samples = m->ctx->frame_size;
						//m->f->nb_samples = 1024;
						m->f->format = m->ctx->sample_fmt;
						m->f->channel_layout = m->ctx->channel_layout;
						m->f->pts =0;

						m->release_frame = 1;
					}

					static void aacenc_release(aacenc_t *m) {
						// release context
						if (m->release_ctx) {
							avcodec_close(m->ctx);
							av_free(m->ctx);
						}

						// release frame
						if (m->release_frame) {
							av_frame_free(&m->f);
						}
					}

					static void bind_buf(aacenc_t *m, uint8_t* samples) {
						int ret = avcodec_fill_audio_frame(m->f, 2, AV_SAMPLE_FMT_S16,(const uint8_t*)samples, 8192, 0);

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
						av_log(m->ctx, AV_LOG_INFO, "got %d size %d, pkt_dts:%ld, frame_dts:%ld\n", m->got, pkt.size,pkt.dts,m->f->pts);
						m->f->pts += 1024;

						m->size = pkt.size;
						m->pts = pkt.pts;
						m->dts = pkt.dts;
					}

					static int aacenc_flush(aacenc_t *m) {
						int ret;

						AVPacket pkt;
						av_init_packet(&pkt);
						pkt.data = m->buf;
						pkt.size = sizeof(m->buf);

						ret = avcodec_encode_audio2(m->ctx, &pkt, NULL, &m->got);

						m->size = pkt.size;
						m->pts = pkt.pts;
						m->dts = pkt.dts;

						return ret;
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
import (
	"fmt"
	"log"
)

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

const (
	AV_CODEC_ID_NONE = C.AV_CODEC_ID_NONE
	AV_CODEC_ID_H264 = C.AV_CODEC_ID_H264
	AV_CODEC_ID_VP8  = C.AV_CODEC_ID_VP8
)

type AChannelSamples []byte

type AACOut struct {
	Data []byte
	Pts  int64
	Dts  int64
	Ts   int64
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

func NewAACEncoderFromCtx(ctx *C.AVCodecContext) (m *AACEncoder) {
	m = &AACEncoder{
		SampleRate:    int(ctx.sample_rate),
		BitRate:       int(ctx.bit_rate),
		Channels:      int(ctx.channels),
		ChannelLayout: int(ctx.channel_layout),
		Profile:       int(ctx.profile),
		sampleSize:    int(C.av_get_bytes_per_sample(ctx.sample_fmt)),
		frameSize:     int(ctx.frame_size),
	}
	m.m.ctx = ctx

	// AV_SAMPLE_FMT_S16
	// m.sampleSize = 2
	// m.frameSize = int(m.m.ctx.frame_size)

	C.aacenc_new_frame(&m.m)

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

func (m *AACEncoder) Release() {
	C.aacenc_release(&m.m)
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

func (m *AACEncoder) samplesToBufMulti(channels []AChannelSamples, samplesCount, samplesOffset int) int {
	mixedSampleSize := m.Channels * m.sampleSize

	// copy samples count
	sz := samplesCount - samplesOffset

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

func (m *AACEncoder) samplesToBufMono(samples AChannelSamples, samplesCount, samplesOffset int) int {
	sz := samplesCount - samplesOffset

	copy(m.bufR[m.bufLen*m.sampleSize:], samples[samplesOffset*m.sampleSize:sz*m.sampleSize])
	m.bufLen += sz

	return sz
}

func (m *AACEncoder) samplesToBuf(channels []AChannelSamples, samplesCount, samplesOffset int) int {
	if len(channels) > 1 {
		return m.samplesToBufMulti(channels, samplesCount, samplesOffset)
	}

	return m.samplesToBufMono(channels[0], samplesCount, samplesOffset)
}

func (m *AACEncoder) HasFrame() bool {
	return m.bufLen >= m.frameSize
}

func (m *AACEncoder) BufLen() int {
	return m.bufLen
}

func (m *AACEncoder) Flush2() (ret *AACOut, got bool, err error) {
	// encode samples
	r := C.aacenc_flush(&m.m)
	if int(r) < 0 {
		err = fmt.Errorf("AACEncoder flush, encode failed")
		return
	}

	// check got flag
	got = int(m.m.got) > 0
	if !got {
		return
	}

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

func (m *AACEncoder) Encode2(channels []AChannelSamples) (ret *AACOut, err error) {
	if channels != nil {
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
		m.samplesToBuf(channels, samplesCount, 0)
	}

	// until codec buf not full, return no_data
	if m.bufLen < m.frameSize {
		err = errors.New("more data")
		return
	}

	// encode samples
	C.aacenc_encode(&m.m)

	// shift buff
	copy(m.bufR[0:], m.bufR[m.frameSize*m.Channels*m.sampleSize:])
	m.bufLen -= m.frameSize

	// check got flag
	if m.m.got == 0 {
		err = errors.New("no data")
		return
	}

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

func (m *AACEncoder) Encode3(channels []AChannelSamples, timeStamp int64) (ret *AACOut, err error) {
	if channels != nil {
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
		m.samplesToBuf(channels, samplesCount, 0)
	}

	// until codec buf not full, return no_data
	if m.bufLen < m.frameSize {
		err = errors.New("more data")
		return
	}

	// set pts, pts in frame count
	// recalc ts from ms to frames
	//m.m.f.pts = C.int64_t(timeStamp * 44100 / 1000)
	//m.m.f.pts = C.int64_t(timeStamp * 48)
	//m.m.f.pts += 1024

	// encode samples
	C.aacenc_encode(&m.m)

	// shift buff
	copy(m.bufR[0:], m.bufR[m.frameSize*m.Channels*m.sampleSize:])
	m.bufLen -= m.frameSize

	// check got flag
	if m.m.got == 0 {
		err = errors.New("no data")
		return
	}

	// extract and return encoded data
	ret = &AACOut{
		Data: make([]byte, (int)(m.m.size)),
		Pts:  int64(m.m.pts),
		Dts:  int64(m.m.dts),
		Ts:   timeStamp,
		//Ts: int64(m.m.pts) * 20 / 1024,
	}

	C.memcpy(
		unsafe.Pointer(&ret.Data[0]),
		unsafe.Pointer(&m.m.buf[0]),
		(C.size_t)(m.m.size),
	)

	return
}
