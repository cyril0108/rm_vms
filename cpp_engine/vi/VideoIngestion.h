#pragma once
#include <string>
#include <thread>
#include <atomic>
#include <memory>
#include <vector>

#include "Log.h"
#include "AVDictionary.h"
#include "SharedMemory.h"
#include "Recording.h"
#include "SafeQueue.h"
#include "AudioTranscoder.h"

// --- Forward Declarations ---
struct AVFormatContext;
struct AVDictionary;
struct AVBSFContext;
struct AVPacket;

class TSMuxer;

struct VideoIngestionConfig {
    std::shared_ptr<ISharedMemory> shm;
    int camID = -1;
    std::string url;
    std::string rootPath;
    std::string profile;
    bool recording = false;
};

class VideoIngestion
{
public:
    VideoIngestion(const VideoIngestionConfig& config);
    ~VideoIngestion();

    void stopIngestion();
    void startRecording();
    void stopRecording();

private:
    std::unique_ptr<RecorderWorker> recorderWorker;
    std::shared_ptr<ISharedMemory> shm;

    VideoIngestionConfig cfg;
    int camID;
    int shmChannelID = -1;
    std::string camName;
    std::string url;
    std::string profile;

    std::string camJsonPartial;
    std::string recStatus;
    std::atomic<bool> recording{false};

    // --- FFmpeg Contexts & Options ---
    AVFormatContext* fmtCtx = nullptr;
    AVDictionary* options = nullptr;
    AVBSFContext* bsfCtx = nullptr;

    AudioTranscoder transcoder;
    TSMuxer* tsMuxer = nullptr;

    // Threading controls
    std::atomic<bool> stopSignal{false};
    std::thread workerThread;
    std::thread diskWriterThread;
    SafeQueue<AVPacket*> diskWriterQueue;

    // --- Stream Tracking ---
    int videoStreamIndex = -1;
    int audioStreamIndex = -1;
    uint32_t videoCodecID = 0;
    uint32_t audioCodecID = 0;
    bool waitForKeyFrame = true;

    int startIngestion();
    int openInput();
    int cleanup();
    void stopAndJoinDiskWriterThread();
    void sendStreamCodecs();
    void updateRECStatus();

    // --- Setup Helpers ---
    void findStreamIndices();
    void initDiskWriter();
    int initVideoFilter();
    bool isAVVC();
    const char* annexbFilterName();

    static int interruptCallback(void* ctx);

    // --- Packet Routing & Processing ---
    void routePacket(AVPacket* packet);
    void ingestVideo(AVPacket* packet);
    void ingestAudio(AVPacket* packet);
    void packetToDiskWriter(AVPacket* packet);
};