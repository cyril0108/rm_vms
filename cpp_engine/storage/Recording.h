#pragma once

#include "SafeQueue.h"
#include "SegmentRecorder.h"

struct AVPacket;
struct AVStream;

// void Recording(AVPacket* packet);

// Spawns the worker loop for multiplexing packets to disk
// void writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID);
// void writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID, const std::string& rootPath = "");

class RecorderWorker {
private:
    SegmentRecorder recorder;

    std::string profile;
    std::string rootPath;
    std::string currentFilePath;
    long currentStartTimeMs = 0;

    void sendSegmentDoneIPC(int camID, long startTimeMs, long endTimeMs, const std::string& filePath);
    long getEndTimeMs(SegmentRecorder& recorder);

    void finalizeCurrentSegment(int camID);

public:
    RecorderWorker(std::string rp = "", std::string prof = "");
    ~RecorderWorker() = default;

    void writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID);

};