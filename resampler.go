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
							#include "libavresample/avresample.h"

							typedef struct {
							    AVAudioResampleContext *avr;

							    int64_t in_sample_rate;
							    int64_t out_sample_rate;
							    int64_t in_sample_fmt;
							    int64_t out_sample_fmt;
							    int64_t in_channel_layout;
							    int64_t out_channel_layout;
								int channels;
							} resampler_t ;

							static int init_resampler(resampler_t *m) {
							    int error;

							    m->avr = avresample_alloc_context();

						        av_log(m->avr, AV_LOG_DEBUG, "in_channel_layout: %ld\n", m->in_channel_layout);
						        av_log(m->avr, AV_LOG_DEBUG, "out_channel_layout: %ld\n", m->out_channel_layout);
						        av_log(m->avr, AV_LOG_DEBUG, "in_sample_fmt: %ld\n", m->in_sample_fmt);
						        av_log(m->avr, AV_LOG_DEBUG, "out_sample_fmt: %ld\n", m->out_sample_fmt);
						        av_log(m->avr, AV_LOG_DEBUG, "in_sample_rate: %ld\n", m->in_sample_rate);
						        av_log(m->avr, AV_LOG_DEBUG, "out_sample_rate: %ld\n", m->out_sample_rate);

							    av_opt_set_int(m->avr, "in_channel_layout",  m->in_channel_layout, 0);
							    av_opt_set_int(m->avr, "out_channel_layout", m->out_channel_layout, 0);
							    av_opt_set_int(m->avr, "in_sample_rate",     m->in_sample_rate, 0);
							    av_opt_set_int(m->avr, "out_sample_rate",    m->out_sample_rate, 0);
							    av_opt_set_int(m->avr, "in_sample_fmt",      m->in_sample_fmt, 0);
							    av_opt_set_int(m->avr, "out_sample_fmt",     m->out_sample_fmt, 0);

							    if ((error = avresample_open(m->avr)) < 0) {
							        av_log(m->avr, AV_LOG_DEBUG, "Could not open resample context\n");
							        avresample_free(&m->avr);

							        return error;
							    }

								m->channels = av_get_channel_layout_nb_channels(m->in_channel_layout);

							    return 0;
							}

				            static int convert_resampler(resampler_t *m, uint8_t *out_data, int out_linesize, int out_samples, uint8_t *in_data, int in_linesize, int in_samples) {
				                int i, plane_size, conv_samples;

				                uint8_t *in_data_array[5];
				                uint8_t *out_data_array[1];

								plane_size = in_linesize / m->channels;
								for (i =0; i < m->channels; ++i) {
									in_data_array[i] = &in_data[i*plane_size];
								}

				                out_data_array[0] = out_data;

		                        av_log(m->avr, AV_LOG_DEBUG, "out_linesize: %d, out_samples: %d, in_linesize: %d, in_samples: %d\n", out_linesize, out_samples, in_linesize, in_samples);

				                conv_samples = avresample_convert(m->avr, &out_data_array[0], out_linesize, out_samples, &in_data_array[0], in_linesize, in_samples);

				                return conv_samples;
				            }

							static int release_resampler(resampler_t *m) {
							    avresample_close(m->avr);
							    avresample_free(&m->avr);

							    return 0;
							}
	*/
	"C"
	"fmt"
	//	"strings"
	//"log"
)
import "unsafe"

type AVResampler struct {
	InSampleRate     int
	OutSampleRate    int
	InSampleFmt      int32
	OutSampleFmt     int32
	InChannelLayout  int
	OutChannelLayout int

	m             C.resampler_t
	convertBuffer []byte
}

func (r *AVResampler) Init() error {
	r.m.in_sample_rate = C.int64_t(r.InSampleRate)
	r.m.out_sample_rate = C.int64_t(r.OutSampleRate)
	r.m.in_sample_fmt = C.int64_t(r.InSampleFmt)
	r.m.out_sample_fmt = C.int64_t(r.OutSampleFmt)
	r.m.in_channel_layout = C.int64_t(r.InChannelLayout)
	r.m.out_channel_layout = C.int64_t(r.OutChannelLayout)

	avLock.Lock()
	defer avLock.Unlock()

	res := C.init_resampler(&r.m)
	if int(res) < 0 {
		return fmt.Errorf("open resampler failed")
	}

	return nil
}

func (r *AVResampler) Convert(samples []byte) ([]byte, error) {
	sampleSize := 2
	if r.InSampleFmt == AV_SAMPLE_FMT_FLT || r.InSampleFmt == AV_SAMPLE_FMT_FLTP {
		sampleSize = 4
	}

	channels := 1
	if r.InChannelLayout == AV_CH_LAYOUT_STEREO {
		channels = 2
	}

	// lineSize - size of all bufer
	inLineSize := len(samples)
	inSamples := inLineSize / channels / sampleSize
	outSamples := C.avresample_get_out_samples(r.m.avr, C.int(inSamples))
	// size(AV_SAMPLE_FMT_S16) = 2
	outLineSize := outSamples * 2

	// create new buf if need
	// AV_CH_LAYOUT_MONO = 1
	convertedSamples := make([]byte, outLineSize*1)

	C.convert_resampler(&r.m, (*C.uint8_t)(unsafe.Pointer(&convertedSamples[0])), C.int(outLineSize), C.int(outSamples),
		(*C.uint8_t)(unsafe.Pointer(&samples[0])), C.int(inLineSize), C.int(inSamples))

	return convertedSamples, nil
}

func (r *AVResampler) Release() error {
	C.release_resampler(&r.m)

	return nil
}
