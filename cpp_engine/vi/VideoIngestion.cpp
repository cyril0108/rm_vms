#include "VideoIngestion.h"

#include <iostream>
// #include <string>
#include <chrono>

#include "utils/Time.h"

extern "C" {
#include <libavformat/avformat.h>
#include <libavcodec/avcodec.h> // You may need this for av_packet_alloc, etc.
}

// --- Constructor ---
VideoIngestion::VideoIngestion(const VideoIngestionConfig& config)
    : shm(config.shm), 
      camID(config.camID), 
      url(config.url),
      profile(config.profile),
      recording(config.recording)
{
    camName = "[Cam" + std::to_string(camID) + "]";
    shmChannelID = shm->ChannelForCamID(camID);
    recorderWorker = std::make_unique<RecorderWorker>(config.rootPath, profile);

    camJsonPartial = "\"cam\":" + std::to_string(camID) + ", \"profile\": \"" + profile + "\"";

    if(shmChannelID < 0) {

        Log::error(camName + "SHM reached max channel!");
        throw std::runtime_error("Failed to allocate SHM channel for camera.");

    } else {

        workerThread = std::thread(&VideoIngestion::startIngestion, this);

    }

}

// --- Destructor ---
VideoIngestion::~VideoIngestion() {

    stopIngestion();

    // Wait for the thread to finish (Join)
    // If we don't join, the thread might try to access 'this' after the object is destroyed -> Crash.
    if (workerThread.joinable()) {
        workerThread.join();
    }

}

/**
 * =========================================================
 * --- Private Method: startIngestion ---
 * =========================================================
 */
int VideoIngestion::startIngestion() {
    avformat_network_init();
    options = configureAVDictionary(nullptr);

    // Connect to Camera
    if (openInput() < 0) return cleanup();
    if (avformat_find_stream_info(fmtCtx, nullptr) < 0) {
        Log::error(camName + " Could not retrieve stream info.");
        return cleanup();
    }

    // Locate Streams
    findStreamIndices();
    if (videoStreamIndex == -1) {
        Log::error(camName + " No video stream found.");
        return cleanup();
    }

    // Setup Filters
    if (initVideoFilter() < 0) return cleanup();

    // ---------------------------------------------------------
    // SPAWN WRITER THREAD HERE
    // Now that we have AVStreams, we can pass codec parameters to the writer
    // ---------------------------------------------------------
    initDiskWriter();

    Log::info(camName + " Connected! Starting Ingestion Loop...");
    Log::send("{\"status\":\"streaming\", " + camJsonPartial + ", \"channel\":" + std::to_string(shmChannelID) + "}");

    // The Main Loop
    AVPacket* packet = av_packet_alloc();
    while (!stopSignal) {
        if (av_read_frame(fmtCtx, packet) < 0) {
            Log::info(camName + " Error or End of Stream.");
            break; // Drop out of loop to trigger reconnect/cleanup
        }

        routePacket(packet);

        // Reset the packet for the next av_read_frame iteration
        av_packet_unref(packet); 
    }

    av_packet_free(&packet);
    return cleanup();
}

/**
 * =========================================================
 * --- Method: Process Control ---
 * =========================================================
 */

void VideoIngestion::stopIngestion() {

    Log::info(camName + " Stop requested by orchestrator...");

    stopSignal = true;

}

void VideoIngestion::stopRecording() {

    Log::info(camName + " Stop recording requested...");

    recording.store(false, std::memory_order_relaxed);

    // Send a "Flush" packet instead of a kill pill
    AVPacket* flushPacket = av_packet_alloc();
    flushPacket->size = 0; 
    diskWriterQueue.push(flushPacket);

}

void VideoIngestion::startRecording() {

    Log::info(camName + " Start recording...");
    recording = true;

}

/**
 * =========================================================
 * Initializations
 * =========================================================
 */
void VideoIngestion::findStreamIndices() {
    videoStreamIndex = -1;
    audioStreamIndex = -1;

    // Parameters: Context, Media Type, Wanted Stream (-1 for auto), Related Stream (-1 for none), Decoder ptr, Flags
    int vIdx = av_find_best_stream(fmtCtx, AVMEDIA_TYPE_VIDEO, -1, -1, nullptr, 0);

    if (vIdx >= 0) {
        videoStreamIndex = vIdx;
        videoCodecID = fmtCtx->streams[vIdx]->codecpar->codec_id;
        Log::info(camName + " Found Video Stream ("+std::to_string(videoCodecID) +") at index: " + std::to_string(videoStreamIndex));
    }

    // We pass 'vIdx' as the related stream so FFmpeg tries to find an audio track explicitly mapped to our video
    int aIdx = av_find_best_stream(fmtCtx, AVMEDIA_TYPE_AUDIO, -1, vIdx, nullptr, 0);
    if (aIdx >= 0) {
        audioStreamIndex = aIdx;
        audioCodecID = fmtCtx->streams[aIdx]->codecpar->codec_id;
        Log::info(camName + " Found Audio Stream ("+std::to_string(audioCodecID) +") at index: " + std::to_string(audioStreamIndex));
    }

}

void VideoIngestion::initDiskWriter() {

    diskWriterThread = std::thread([this]() {
        AVStream* vStream = (videoStreamIndex != -1) ? fmtCtx->streams[videoStreamIndex] : nullptr;
        AVStream* aStream = (audioStreamIndex != -1) ? fmtCtx->streams[audioStreamIndex] : nullptr;

        // Pass both streams to the worker
        this->recorderWorker->writerWorker(this->diskWriterQueue, vStream, aStream, this->camID);
    });

}

const char* VideoIngestion::annexbFilterName() {
    if (videoCodecID == AV_CODEC_ID_H264) {
        return "h264_mp4toannexb";
    } else if (videoCodecID == AV_CODEC_ID_HEVC) {
        return "hevc_mp4toannexb";
    }
    return nullptr;
}

// If it's AVCC, it needs length-prefixes converted AND SPS/PPS injected.
bool VideoIngestion::isAVVC() {
    AVStream* vStream = fmtCtx->streams[videoStreamIndex];
    uint8_t* extradata = vStream->codecpar->extradata;
    int extradata_size = vStream->codecpar->extradata_size;

    return extradata_size > 0 && extradata[0] == 1;
}

int VideoIngestion::initVideoFilter() {

    const char* filterName = nullptr;

    if(isAVVC()) {
        filterName = annexbFilterName();
        if (!filterName) {
            Log::error(camName + " Detected AVCC, but codec is not H264/HEVC. Cannot apply BSF.");
            return -1;
        }
        Log::info(camName + " Detected AVCC stream. Applying " + std::string(filterName) + ".");
    } else {
        // Annex-B format detected (or no extradata). Start codes are present.
        // We only need to ensure SPS/PPS is injected into the stream.
        Log::info(camName + " Detected Annex-B stream. Applying dump_extra.");
        filterName = "dump_extra";
    }

    const AVBitStreamFilter *bsf = av_bsf_get_by_name(filterName);

    if (av_bsf_alloc(bsf, &bsfCtx) < 0) {
        Log::error(camName + " Failed to allocate " + filterName + "BSF.");
        return -1;
    }

    if (avcodec_parameters_copy(bsfCtx->par_in, fmtCtx->streams[videoStreamIndex]->codecpar) < 0) {
        Log::error(camName + " Failed to copy parameters to BSF.");
        return -1;
    }

    if (av_bsf_init(bsfCtx) < 0) {
        Log::error(camName + " Failed to initialize BSF.");
        return -1;
    }

    return 0;
}

void VideoIngestion::routePacket(AVPacket* packet) {
    if (packet->stream_index == videoStreamIndex) {
        ingestVideo(packet);
    } else if (packet->stream_index == audioStreamIndex) {
        ingestAudio(packet);
    } 
    // If it's metadata or subtitles, we just do nothing.
    // The orchestrator loop will safely unref it.
}

/**
 * =========================================================
 * Disk Writer
 * =========================================================
 */

void VideoIngestion::packetToDiskWriter(AVPacket* packet) {

    if (!recording) {
        return;
    }

    AVPacket* cloneForDisk = av_packet_alloc();
    if (av_packet_ref(cloneForDisk, packet) >= 0) {

        // If the queue rejects the packet, WE must free the clone.
        if (!diskWriterQueue.push(cloneForDisk)) {
            av_packet_unref(cloneForDisk);
            av_packet_free(&cloneForDisk);
        }

    } else {
        av_packet_free(&cloneForDisk);
        Log::error(camName + "Failed to ref-count packet for disk queue.");
    }

}

/**
 * =========================================================
 * Ingestion
 * =========================================================
 */
FrameMetadata VideoIngestion::makeFrameMetadataV(AVPacket* packet, bool isKey) {
    FrameMetadata meta;

    // Grab the absolute Wall-Clock time in milliseconds
    meta.epochMs = utils::getCurrentEpochMSTime();

    meta.magic = WrapMagicNumber;
    meta.frameSize = packet->size;
    meta.pts = packet->pts;
    meta.dts = packet->dts;
    meta.isKeyFrame = isKey ? 1 : 0;
    meta.codecID = videoCodecID;
    meta.mediaType = static_cast<uint8_t>(MediaType::VIDEO);

    return meta;

}

FrameMetadata VideoIngestion::makeFrameMetadataA(AVPacket* packet) {
    // Construct the metadata cleanly
    FrameMetadata meta;
    meta.epochMs = utils::getCurrentEpochMSTime();

    meta.magic = WrapMagicNumber;
    meta.frameSize = packet->size;
    meta.pts = packet->pts;
    meta.dts = packet->dts;
    meta.isKeyFrame = 0;
    meta.codecID = audioCodecID;
    meta.mediaType = static_cast<uint8_t>(MediaType::AUDIO);
    return meta;
}

void VideoIngestion::ingestVideo(AVPacket* packet) {
    // Send raw packet to the filter
    if (av_bsf_send_packet(bsfCtx, packet) == 0) {

        // Allocate a temporary packet for the output
        AVPacket* bsfPacket = av_packet_alloc();

        // Receive the modified packet(s) into the NEW packet
        while (av_bsf_receive_packet(bsfCtx, bsfPacket) == 0) {

            bool isKey = (bsfPacket->flags & AV_PKT_FLAG_KEY);

            if (waitForKeyFrame) {
                if (isKey) {
                    Log::info(camName + " [ingestVideo] First key frame found.");
                    waitForKeyFrame = false;
                } else {
                    av_packet_unref(bsfPacket); 
                    continue;
                }
            }

            try {

                FrameMetadata meta = makeFrameMetadataV(bsfPacket, isKey);

                // Use bsfPacket here!
                if (shm->WriteFrame(shmChannelID, meta, bsfPacket->data) < 0) {
                    Log::error(camName + " [ingestVideo] Failed to write frame data.");
                }
                packetToDiskWriter(bsfPacket);

            } catch(...) {
                Log::error(camName + " [ingestVideo] Caught exception writing frame data.");
            }

            // Clean up the temporary packet for the next iteration of the while loop
            av_packet_unref(bsfPacket);
        }

        // Free the temporary packet when the while loop is done
        av_packet_free(&bsfPacket);
    }
}

void VideoIngestion::ingestAudio(AVPacket* packet) {

    // A/V Sync Gatekeeper: Drop audio until video has established a keyframe.
    // This prevents "black screen with audio" at the start of recordings/streams.
    if (waitForKeyFrame) {
        return;
    }

    // Audio frames do not have "Keyframes" in the same way video does, 
    // so we pass 'false' for the isKey parameter.
    try {
        FrameMetadata meta = makeFrameMetadataA(packet);
        if (shm->WriteFrame(shmChannelID, meta, packet->data) < 0) {
            Log::error(camName + " [SHM] Failed to write audio stream.");
        }
    } catch(...) {
        Log::error(camName + " [SHM] Caught exception writing audio stream.");
    }

    packetToDiskWriter(packet);

}


// --- Private Method: openInput ---
/**
 * Opens url and handles error message
 * @return 0 on success, -1 on failure.
 */
int VideoIngestion::openInput() {

    Log::info(camName + "Connecting to: " + url);

    fmtCtx = avformat_alloc_context();
    if (!fmtCtx) {
        Log::error(camName + " Failed to allocate AVFormatContext.");
        return -1;
    }

    fmtCtx->interrupt_callback.callback = VideoIngestion::interruptCallback;
    fmtCtx->interrupt_callback.opaque = this;

    int ret = avformat_open_input(&fmtCtx, url.c_str(), nullptr, &options);
    if (ret != 0) {
        // Create a buffer for the error message
        char errbuf[256];

        // Ask FFmpeg to translate the error code
        av_strerror(ret, errbuf, sizeof(errbuf));

        std::cerr << camName << "[FFmpeg Error] Could not open source: " << url << std::endl;
        std::cerr << "Reason: " << errbuf << " (Code: " << ret << ")" << std::endl;

        // Note: avformat_open_input automatically frees fmtCtx on failure, 
        // so we must set it back to nullptr to prevent double-free in cleanup()
        fmtCtx = nullptr;

        return -1;
    }

    return 0;

}

/**
 * =========================================================
 * --- Private Method: stopAndJoinDiskWriterThread ---
 * =========================================================
 */
void VideoIngestion::stopAndJoinDiskWriterThread() {

    // Wake up the disk writer thread and tell it to exit safely
    diskWriterQueue.push(nullptr);

    // Join the disk writer thread
    if (diskWriterThread.joinable()) {
        diskWriterThread.join();
    }
}

/**
 * =========================================================
 * --- FFmpeg Interrupt Callback ---
 * =========================================================
 */
int VideoIngestion::interruptCallback(void* ctx) {
    // Cast the opaque pointer back to our class instance
    VideoIngestion* vi = static_cast<VideoIngestion*>(ctx);

    // If the instance exists and the stop signal is true, return 1 to abort!
    if (vi && vi->stopSignal.load(std::memory_order_relaxed)) {
        return 1;
    }

    // Return 0 to let FFmpeg continue blocking/reading
    return 0;
}

/**
 * =========================================================
 * --- Private Method: cleanup ---
 * =========================================================
 */
int VideoIngestion::cleanup() {

    stopAndJoinDiskWriterThread();

    if (shmChannelID >= 0) {
        shm->ReleaseChannelForCamID(camID);
        shmChannelID = -1; // Reset to prevent double-release if cleanup is called twice
    }

    // Free the Bitstream Filter (Fixes the memory leak!)
    if (bsfCtx) {
        av_bsf_free(&bsfCtx);
        bsfCtx = nullptr;
    }

    // Close the input and free context
    if (fmtCtx) {
        avformat_close_input(&fmtCtx); 
        fmtCtx = nullptr;
    }

    // Free the dictionary options
    if (options) {
        av_dict_free(&options);
        options = nullptr;
    }

    // De-initialize network
    avformat_network_deinit();

    Log::info(camName + " Thread Exited cleanly.");
    Log::send("{\"status\":\"stopped\", " + camJsonPartial + "}");

    return -1; // Or return 0 depending on how your worker thread monitors exits
}