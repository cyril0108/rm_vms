#pragma once
#include <string>
#include <thread>
#include <atomic>

#include "Log.h"
#include "AVDictionary.h"
#include "SharedMemory.h"
#include "Recording.h"
#include "SafeQueue.h"

// --- Forward Declarations ---
// Tell the compiler these are structs, which is all it needs to create pointers.
// This completely removes the FFmpeg dependency from the header file.
struct AVFormatContext;
struct AVDictionary;
struct AVBSFContext;
struct AVPacket;

class VideoIngestion
{
public:
    VideoIngestion(std::shared_ptr<ISharedMemory> mm, int id, const std::string u, const std::string rp);
    ~VideoIngestion();
    // stop
    void stopIngestion();

private:

    std::unique_ptr<RecorderWorker> recorderWorker;
    std::shared_ptr<ISharedMemory> shm;

    int camID;
    int shmChannelID = -1;
    std::string camName;
    std::string url;

    // --- FFmpeg Contexts & Options ---
    AVFormatContext* fmtCtx = nullptr;
    AVDictionary* options = nullptr;
    AVBSFContext* bsfCtx = nullptr;       // Bitstream filter context for SPS/PPS injection

    // Threading controls
    std::atomic<bool> stopSignal{false}; // Each camera has its own stop flag
    std::thread workerThread;            // The thread handling the loop
    std::thread diskWriterThread;
    SafeQueue<AVPacket*> diskWriterQueue;

    // --- Stream Tracking ---
    int videoStreamIndex = -1;
    int audioStreamIndex = -1;
    uint32_t videoCodecID = 0;
    uint32_t audioCodecID = 0;
    bool waitForKeyFrame = true;       // Ensures we drop P-frames until our first IDR

    int startIngestion();       // The main worker thread loop
    int openInput();            // Connects to the RTSP source
    int cleanup();              // Safely frees all FFmpeg resources
    void stopAndJoinDiskWriterThread();

    // --- Setup Helpers ---
    void findStreamIndices();
    void initDiskWriter();
    int initVideoFilter();
    FrameMetadata makeFrameMetadataV(AVPacket* packet, bool isKey);
    FrameMetadata makeFrameMetadataA(AVPacket* packet);

    // FFMPEG interrupt callback
    static int interruptCallback(void* ctx);

    // --- Packet Routing & Processing ---
    void routePacket(AVPacket* packet);
    void ingestVideo(AVPacket* packet);
    void ingestAudio(AVPacket* packet);
    void packetToDiskWriter(AVPacket* packet);
};
