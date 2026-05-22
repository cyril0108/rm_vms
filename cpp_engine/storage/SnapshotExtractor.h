#pragma once

#include <string>

extern "C" {
#include <libavcodec/avcodec.h>
#include <libswscale/swscale.h>
#include <libavutil/imgutils.h>
}

class SnapshotExtractor {
private:
    AVCodecContext* decoderCtx = nullptr;
    AVCodecContext* encoderCtx = nullptr;
    SwsContext* swsCtx = nullptr;

    AVFrame* rawFrame = nullptr;
    AVFrame* scaledFrame = nullptr;
    AVPacket* jpegPacket = nullptr;

    bool isInitialized = false;

    void cleanup();

public:
    SnapshotExtractor() = default;
    ~SnapshotExtractor();

    // Initializes the H.264/H.265 decoder and the MJPEG encoder
    bool Initialize(AVCodecParameters* inVideoCodecPar, int outWidth, int outHeight);
    
    // Decodes the keyframe, scales it, encodes to JPG, and saves to disk
    bool ExtractAndSave(AVPacket* keyframePacket, const std::string& outFilePath);
};