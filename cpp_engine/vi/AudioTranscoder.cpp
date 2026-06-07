#include "AudioTranscoder.h"

extern "C" {
#include <libavformat/avformat.h>
}

AudioTranscoder::~AudioTranscoder() {
    if (decCtx) avcodec_free_context(&decCtx);
    if (encCtx) avcodec_free_context(&encCtx);
    if (swrCtx) swr_free(&swrCtx);
    if (fifo) av_audio_fifo_free(fifo);
}

bool AudioTranscoder::init(AVStream* inStream) {
    // Setup Decoder
    const AVCodec* dec = avcodec_find_decoder(inStream->codecpar->codec_id);
    if (!dec) return false;

    decCtx = avcodec_alloc_context3(dec);
    avcodec_parameters_to_context(decCtx, inStream->codecpar);

    // --- CRITICAL FIX: FORCE CHANNEL LAYOUT IF CAMERA OMITTED IT ---
    #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
        if (decCtx->ch_layout.order == AV_CHANNEL_ORDER_UNSPEC || decCtx->ch_layout.nb_channels == 0) {
            int channels = decCtx->ch_layout.nb_channels > 0 ? decCtx->ch_layout.nb_channels : 1;
            av_channel_layout_default(&decCtx->ch_layout, channels);
        }
    #else
        if (decCtx->channel_layout == 0) {
            int channels = decCtx->channels > 0 ? decCtx->channels : 1;
            decCtx->channel_layout = av_get_default_channel_layout(channels);
            decCtx->channels = channels;
        }
    #endif

    if (avcodec_open2(decCtx, dec, nullptr) < 0) return false;

    // Setup 48kHz AAC Encoder
    const AVCodec* enc = avcodec_find_encoder(AV_CODEC_ID_AAC);
    encCtx = avcodec_alloc_context3(enc);
    encCtx->sample_rate = 48000;          // Standard Broadcast Frequency
    encCtx->sample_fmt = enc->sample_fmts[0]; // Usually FLTP
    encCtx->bit_rate = 128000;
    
    #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
        av_channel_layout_default(&encCtx->ch_layout, 2); // Force Stereo
    #else
        encCtx->channel_layout = AV_CH_LAYOUT_STEREO;
        encCtx->channels = 2;
    #endif

    if (avcodec_open2(encCtx, enc, nullptr) < 0) return false;

    // Setup Resampler (Camera Layout -> 48kHz Stereo)
    swrCtx = swr_alloc();
    #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
        swr_alloc_set_opts2(&swrCtx, 
            &encCtx->ch_layout, encCtx->sample_fmt, encCtx->sample_rate,
            &decCtx->ch_layout, decCtx->sample_fmt, decCtx->sample_rate,
            0, nullptr);
    #else
        swr_alloc_set_opts(swrCtx,
            encCtx->channel_layout, encCtx->sample_fmt, encCtx->sample_rate,
            decCtx->channel_layout, decCtx->sample_fmt, decCtx->sample_rate,
            0, nullptr);
    #endif
    
    // --- CRITICAL FIX: PREVENT SEGFAULT BY CHECKING SWR_INIT ---
    if (swr_init(swrCtx) < 0) {
        // If the resampler STILL fails to initialize, gracefully abort 
        // the transcoder so it passes raw audio instead of crashing.
        return false;
    }

    // Setup FIFO buffer (AAC requires strictly 1024 samples per frame)
    fifo = av_audio_fifo_alloc(encCtx->sample_fmt, encCtx->channels, 1);
    isInitialized = true;
    return true;
}

void AudioTranscoder::process(AVPacket* inPkt, std::function<void(AVPacket*)> onEncoded) {
    if (!isInitialized) return;

    // Send to Decoder
    if (avcodec_send_packet(decCtx, inPkt) < 0) return;
    
    AVFrame* decFrame = av_frame_alloc();
    while (avcodec_receive_frame(decCtx, decFrame) == 0) {

        // Allocate Resampled Frame
        AVFrame* resampledFrame = av_frame_alloc();
        resampledFrame->format = encCtx->sample_fmt;
        #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
            av_channel_layout_copy(&resampledFrame->ch_layout, &encCtx->ch_layout);
        #else
            resampledFrame->channel_layout = encCtx->channel_layout;
            resampledFrame->channels = encCtx->channels;
        #endif
        
        resampledFrame->nb_samples = swr_get_out_samples(swrCtx, decFrame->nb_samples);
        av_frame_get_buffer(resampledFrame, 0);

        // Resample Audio
        int outSamples = swr_convert(swrCtx, resampledFrame->data, resampledFrame->nb_samples, 
                                     (const uint8_t**)decFrame->data, decFrame->nb_samples);

        // Push to FIFO
        av_audio_fifo_write(fifo, (void**)resampledFrame->data, outSamples);
        av_frame_free(&resampledFrame);
    }
    av_frame_free(&decFrame);

    // Pull exactly 1024 samples from FIFO and Encode
    AVFrame* encFrame = av_frame_alloc();
    encFrame->nb_samples = encCtx->frame_size;
    encFrame->format = encCtx->sample_fmt;
    #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
        av_channel_layout_copy(&encFrame->ch_layout, &encCtx->ch_layout);
    #else
        encFrame->channel_layout = encCtx->channel_layout;
        encFrame->channels = encCtx->channels;
    #endif
    av_frame_get_buffer(encFrame, 0);

    while (av_audio_fifo_size(fifo) >= encCtx->frame_size) {
        av_audio_fifo_read(fifo, (void**)encFrame->data, encCtx->frame_size);
        encFrame->pts = outPts;
        outPts += encFrame->nb_samples;

        if (avcodec_send_frame(encCtx, encFrame) == 0) {
            AVPacket* outPkt = av_packet_alloc();
            while (avcodec_receive_packet(encCtx, outPkt) == 0) {
                onEncoded(outPkt); // Fire Callback
                av_packet_unref(outPkt);
            }
            av_packet_free(&outPkt);
        }
    }
    av_frame_free(&encFrame);
}