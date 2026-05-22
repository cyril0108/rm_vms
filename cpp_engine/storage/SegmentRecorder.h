#pragma once

#include <string>

extern "C" {
#include <libavformat/avformat.h>
}

class SegmentRecorder {
private:
    AVFormatContext* outFormatCtx = nullptr;
    std::string currentFilename;
    bool isRecording = false;

    // Output stream indices in the MP4 file
    int outVideoStreamIndex = -1;
    int outAudioStreamIndex = -1;

    // Original input stream indices from the camera
    int inVideoStreamIndex = -1;
    int inAudioStreamIndex = -1;

    // Timebases for proper PTS/DTS rescaling
    AVRational videoInputTimeBase;
    AVRational audioInputTimeBase;

    // Time tracking state for sanitization
    int64_t lastVideoDTS = AV_NOPTS_VALUE;
    int64_t lastAudioDTS = AV_NOPTS_VALUE;

    // For tracking actual segment duration
    int64_t firstVideoPTS = AV_NOPTS_VALUE;
    int64_t lastVideoPTS = AV_NOPTS_VALUE;

    // Timeline Normalization Trackers
    bool hasStartTime = false;
    bool hasAudioStartTime = false;
    int64_t startVideoTime = AV_NOPTS_VALUE;
    int64_t startAudioTime = AV_NOPTS_VALUE;

    // Adjusts packets to start at 0. Returns false if the packet should be dropped.
    bool normalizeTimeline(AVPacket* packet); 

    // Forces timestamps to be strictly increasing
    void sanitizeTimestamps(AVPacket* packet, int64_t* lastDTS);

public:
    SegmentRecorder() = default;
    ~SegmentRecorder();

    // Now accepts both streams. Audio can be nullptr if the camera doesn't have a mic.
    bool StartSegment(const std::string& filename, AVStream* inVideoStream, AVStream* inAudioStream);
    void StopSegment();

    void WritePacket(AVPacket* packet);
    double GetVideoDurationMilliseconds() const;

    inline bool IsRecording() const { return isRecording; }
};