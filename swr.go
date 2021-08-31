package gmf

/*

#cgo pkg-config: libswresample

#include "libswresample/swresample.h"
#include <libavcodec/avcodec.h>
#include <libavutil/frame.h>

int gmf_sw_resample(SwrContext* ctx, AVFrame* dstFrame, AVFrame* srcFrame){
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		(const uint8_t **)srcFrame->data, srcFrame->nb_samples);
}

int gmf_swr_flush(SwrContext* ctx, AVFrame* dstFrame) {
	return swr_convert(ctx, dstFrame->data, dstFrame->nb_samples,
		NULL, 0);
}

*/
import "C"

import (
	"fmt"
)

type SwrCtx struct {
	swrCtx          *C.struct_SwrContext
	outChannels     int
	outSampleFormat int32
	outSampleRate   int
	inSampleRate    int
}

func NewSwrCtx(
	outChannelLayout int,
	outSampleFormat int32,
	outSampleRate int,
	inChannelLayout int,
	inSampleFormat int32,
	inSampleRate int,
) (*SwrCtx, error) {
	ctx := &SwrCtx{
		swrCtx: C.swr_alloc_set_opts(
			nil,
			C.int64_t(outChannelLayout),
			outSampleFormat,
			C.int(outSampleRate),
			C.int64_t(inChannelLayout),
			inSampleFormat,
			C.int(inSampleRate),
			0,
			nil,
		),
		outChannels:     int(C.av_get_channel_layout_nb_channels(C.uint64_t(outChannelLayout))),
		outSampleFormat: outSampleFormat,
		outSampleRate:   outSampleRate,
		inSampleRate:    inSampleRate,
	}

	if ret := int(C.swr_init(ctx.swrCtx)); ret < 0 {
		return nil, fmt.Errorf("error initializing swr context - %s", AvError(ret))
	}

	return ctx, nil
}

func (ctx *SwrCtx) Free() {
	C.swr_free(&ctx.swrCtx)
}

func (ctx *SwrCtx) Resample(input *Frame, outSamples int) (*Frame, error) {
	inSamples := input.NbSamples()
	// outSamples := int(C.av_rescale_rnd(C.swr_get_delay(ctx.swrCtx, C.longlong(ctx.inSampleRate))+C.longlong(inSamples), C.longlong(ctx.outSampleRate), C.longlong(ctx.inSampleRate), C.AV_ROUND_UP))

	out, err := NewAudioFrame(ctx.outSampleFormat, ctx.outChannels, outSamples)
	if err != nil {
		return nil, fmt.Errorf("error creating new audio frame - %s\n", err)
	}

	ret := int(C.swr_convert(ctx.swrCtx, (**C.uchar)(&out.avFrame.data[0]), C.int(outSamples), (**C.uchar)(&input.avFrame.data[0]), C.int(inSamples)))
	if ret < 0 {
		out.Free()
		return nil, AvError(ret)
	}

	return out, nil
}
