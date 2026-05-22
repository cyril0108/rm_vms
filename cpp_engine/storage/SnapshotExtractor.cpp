#include "SnapshotExtractor.h"
#include "Log.h"
#include <fstream>

SnapshotExtractor::~SnapshotExtractor() {
    cleanup();
}

void SnapshotExtractor::cleanup() {
    if (decoderCtx) avcodec_free_context(&decoderCtx);
    if (encoderCtx) avcodec_free_context(&encoderCtx);
    if (swsCtx) sws_freeContext(swsCtx);
    if (rawFrame) av_frame_free(&rawFrame);
    if (scaledFrame) av_frame_free(&scaledFrame);
    if (jpegPacket) av_packet_free(&jpegPacket);
    isInitialized = false;
}

bool SnapshotExtractor::Initialize(AVCodecParameters* inVideoCodecPar, int outWidth, int outHeight) {
    cleanup();

    // Setup Video Decoder (H.264 / H.265)
    const AVCodec* decoder = avcodec_find_decoder(inVideoCodecPar->codec_id);
    if (!decoder) return false;
    
    decoderCtx = avcodec_alloc_context3(decoder);
    avcodec_parameters_to_context(decoderCtx, inVideoCodecPar);
    if (avcodec_open2(decoderCtx, decoder, nullptr) < 0) return false;

    // Setup JPEG Encoder
    const AVCodec* encoder = avcodec_find_encoder(AV_CODEC_ID_MJPEG);
    if (!encoder) return false;
    
    encoderCtx = avcodec_alloc_context3(encoder);
    encoderCtx->pix_fmt = AV_PIX_FMT_YUVJ420P; // Standard for MJPEG
    encoderCtx->width = outWidth;
    encoderCtx->height = outHeight;
    encoderCtx->time_base = {1, 25};
    if (avcodec_open2(encoderCtx, encoder, nullptr) < 0) return false;

    // Setup Scaler (Converts camera's pixel format to MJPEG's pixel format and resizes)
    swsCtx = sws_getContext(
        decoderCtx->width, decoderCtx->height, decoderCtx->pix_fmt,
        encoderCtx->width, encoderCtx->height, encoderCtx->pix_fmt,
        SWS_BILINEAR, nullptr, nullptr, nullptr
    );

    // Allocate Memory Buffers
    rawFrame = av_frame_alloc();
    scaledFrame = av_frame_alloc();
    scaledFrame->format = encoderCtx->pix_fmt;
    scaledFrame->width = encoderCtx->width;
    scaledFrame->height = encoderCtx->height;
    av_frame_get_buffer(scaledFrame, 32);
    
    jpegPacket = av_packet_alloc();

    isInitialized = true;
    return true;
}

bool SnapshotExtractor::ExtractAndSave(AVPacket* keyframePacket, const std::string& outFilePath) {
    if (!isInitialized) return false;

    // Decode the packet into raw pixels
    if (avcodec_send_packet(decoderCtx, keyframePacket) < 0) return false;
    if (avcodec_receive_frame(decoderCtx, rawFrame) < 0) return false;

    // Scale and convert pixel format
    sws_scale(swsCtx, rawFrame->data, rawFrame->linesize, 0, decoderCtx->height, 
              scaledFrame->data, scaledFrame->linesize);

    // Encode to JPEG
    if (avcodec_send_frame(encoderCtx, scaledFrame) < 0) return false;
    if (avcodec_receive_packet(encoderCtx, jpegPacket) < 0) return false;

    // Write raw bytes to disk
    std::ofstream outFile(outFilePath, std::ios::binary);
    if (!outFile) return false;
    outFile.write(reinterpret_cast<char*>(jpegPacket->data), jpegPacket->size);
    outFile.close();

    // Clean up the packet for the next run
    av_packet_unref(jpegPacket);
    return true;
}