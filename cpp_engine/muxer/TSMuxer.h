#pragma once
#include "SharedMemory.h"
#include "Log.h"
#include <memory>

extern "C" {
#include <libavformat/avformat.h>
#include <libavutil/avutil.h>
}

class TSMuxer {
public:
    TSMuxer(std::shared_ptr<ISharedMemory> shm, int shmChannelID);
    ~TSMuxer();

    // Initializes the muxer using explicit parameters and timebases
    bool init(AVCodecParameters* vPar, AVRational vTb, 
              AVCodecParameters* aPar, AVRational aTb);

    // Explicit routing methods
    bool muxVideoPacket(AVPacket* pkt, bool isKey);
    bool muxAudioPacket(AVPacket* pkt);

private:
    std::shared_ptr<ISharedMemory> shm;
    int shmChannelID;

    AVFormatContext* outCtx = nullptr;
    AVIOContext* avioCtx = nullptr;
    uint8_t* avioBuffer = nullptr;
    // const int avioBufferSize = 4096; // 32KB chunks
    const int avioBufferSize = 32768; // 32KB chunks

    int outVideoStreamIndex = -1;
    int outAudioStreamIndex = -1;

    bool isKeyFrame = false;

    // We store the input timebases so we can rescale timestamps mathematically
    AVRational inVideoTimeBase;
    AVRational inAudioTimeBase;

    int64_t lastVideoDTS = AV_NOPTS_VALUE;
    int64_t lastAudioDTS = AV_NOPTS_VALUE;

    // --- DIAGNOSTIC COUNTERS ---
    int64_t totalPacketsReceived = 0;
    int64_t totalCallbacksTriggered = 0;
    int64_t totalBytesMuxed = 0;

    void enforceMonotonicity(AVPacket* pkt, int64_t& lastDTSTracker);

    // The magic callback that intercepts FFmpeg's disk writes
    static int shmWriteCallback(void* opaque, uint8_t* buf, int buf_size);
};