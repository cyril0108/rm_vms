#include "VideoIngestion.h"
#include "TSMuxer.h"

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
    updateRECStatus();

    if(shmChannelID < 0) {

        Log::error(camName + "SHM reached max channel!");
        throw std::runtime_error("Failed to allocate SHM channel for camera.");

    } else {

        Log::send("{\"status\":\"shm_id\", " + camJsonPartial + ", \"channel\":" + std::to_string(shmChannelID) + "}");
        workerThread = std::thread(&VideoIngestion::startIngestion, this);

    }

}

void VideoIngestion::updateRECStatus() {
    recStatus = recording ? "recording" : "streaming";
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

    // Extract exact Video parameters
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

    // ---------------------------------------------
    // --- Initialize Universal Audio Transcoder ---
    // ---------------------------------------------
    // If the stream is AAC, OR any PCM telecom codec, we intercept it!
    bool needsTranscoding = (audioStreamIndex >= 0 && 
                            (audioCodecID == AV_CODEC_ID_AAC || 
                             audioCodecID == AV_CODEC_ID_PCM_MULAW || 
                             audioCodecID == AV_CODEC_ID_PCM_ALAW));

    if (needsTranscoding) {
        transcoder.init(fmtCtx->streams[audioStreamIndex]);
    }
    // ---------------------------------------------

    // ---------------------------------------------
    // --- Initialize the MPEG-TS Muxer ------------
    // ---------------------------------------------
    tsMuxer = new TSMuxer(shm, shmChannelID);

    // Extract exact Video parameters
    AVCodecParameters* vPar = (videoStreamIndex >= 0) ? fmtCtx->streams[videoStreamIndex]->codecpar : nullptr;
    AVRational vTb = (videoStreamIndex >= 0) ? fmtCtx->streams[videoStreamIndex]->time_base : AVRational{1, 90000};

    // Extract exact Audio parameters (Prioritize Transcoder output if active)
    AVCodecParameters* aPar = nullptr;
    AVRational aTb = {1, 90000};

    if (transcoder.isInitialized) {
        aPar = avcodec_parameters_alloc();
        avcodec_parameters_from_context(aPar, transcoder.encCtx);
        aTb = {1, transcoder.encCtx->sample_rate};
    } else if (audioStreamIndex >= 0) {
        aPar = fmtCtx->streams[audioStreamIndex]->codecpar;
        aTb = fmtCtx->streams[audioStreamIndex]->time_base;
    }

    // Bind parameters to the Muxer
    tsMuxer->init(vPar, vTb, aPar, aTb);

    if (transcoder.isInitialized && aPar) {
        avcodec_parameters_free(&aPar); // Free temp alloc
    }
    // ---------------------------------------------


    // ---------------------------------------------------------
    // SPAWN WRITER THREAD HERE
    // Now that we have AVStreams, we can pass codec parameters to the writer
    // ---------------------------------------------------------
    initDiskWriter();

    Log::info(camName + " Connected! Starting Ingestion Loop...");
    Log::send("{\"status\":\"" + recStatus +"\", " + camJsonPartial + "}");

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
    updateRECStatus();

    // Send a "Flush" packet instead of a kill pill
    AVPacket* flushPacket = av_packet_alloc();
    flushPacket->size = 0; 
    diskWriterQueue.push(flushPacket);

    // Log::send("{\"status\":\"streaming\", " + camJsonPartial + "}");

}

void VideoIngestion::startRecording() {

    Log::info(camName + " Start recording...");
    recording.store(true, std::memory_order_relaxed);
    updateRECStatus();

    Log::send("{\"status\":\"" + recStatus + "\", " + camJsonPartial + "}");

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

    sendStreamCodecs();

}

void VideoIngestion::sendStreamCodecs() {

    Log::send("{\"status\":\"codecs\", " + camJsonPartial +
       ", \"vcodec\":" + std::to_string(videoCodecID) +
       ", \"acodec\":" + std::to_string(audioCodecID) + "}");

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
 * Timestamp Normalization (SRP Extracted)
 * =========================================================
 */
// void VideoIngestion::normalizeTimestamps(int64_t& pts, int64_t& dts, int streamIndex) {

//     // // --- ACCURATE DYNAMIC TIMEBASE CONVERSION ---
//     // AVRational target_tb = {1, 90000};
//     // AVRational stream_tb = fmtCtx->streams[audioStreamIndex]->time_base;

//     // pts = av_rescale_q(pts, stream_tb, target_tb);
//     // dts = av_rescale_q(dts, stream_tb, target_tb);

//     // ---------------------------------------------
//     // If FFmpeg strips the timestamp, NEVER use the Unix Epoch.
//     // Simply increment safely from the last known good frame.
//     // if (pts == AV_NOPTS_VALUE && dts == AV_NOPTS_VALUE) {
//     //     pts = lastValidPTS + 3000;
//     //     dts = pts;
//     //     return;
//     // }

//     if (pts == AV_NOPTS_VALUE) pts = dts;
//     if (dts == AV_NOPTS_VALUE) dts = pts;

//     // Exact mathematical rescale to 90kHz
//     AVRational target_tb = {1, 90000}; 
//     AVRational stream_tb = fmtCtx->streams[streamIndex]->time_base;

//     pts = av_rescale_q(pts, stream_tb, target_tb);
//     dts = av_rescale_q(dts, stream_tb, target_tb);

//     // Keep track of the highest valid PTS to prevent backward jumping on missing packets
//     lastValidPTS = std::max(lastValidPTS, pts);
// }

/**
 * =========================================================
 * ADTS Audio Helpers (SRP Extracted)
 * =========================================================
 */
// int VideoIngestion::get_adts_sr_index(int sample_rate) {
//     switch(sample_rate) {
//         case 96000: return 0; case 88200: return 1; case 64000: return 2;
//         case 48000: return 3; case 44100: return 4; case 32000: return 5;
//         case 24000: return 6; case 22050: return 7; case 16000: return 8;
//         case 12000: return 9; case 11025: return 10; case 8000: return 11;
//         case 7350: return 12; default: return 4; // default 44100
//     }
// }

// bool VideoIngestion::injectADTS(AVPacket* packet, std::vector<uint8_t>& adtsPayload) {
//     if (audioCodecID != AV_CODEC_ID_AAC) return false;

//     // Check if camera already provided ADTS
//     if (packet->size >= 2 && packet->data[0] == 0xFF && (packet->data[1] & 0xF0) == 0xF0) {
//         return false; 
//     }

//     AVCodecParameters* par = fmtCtx->streams[audioStreamIndex]->codecpar;
    
//     int profile = 1; // Fallback AAC-LC
//     int sr_index = 4; // Fallback 44100Hz
//     int channels = 1;

//     // --- THE FIX: Extract exact parameters from the Camera's AudioSpecificConfig ---
//     if (par->extradata != nullptr && par->extradata_size >= 2) {
//         uint16_t asc = (par->extradata[0] << 8) | par->extradata[1];
//         profile = ((asc >> 11) & 0x1F) - 1; // AOT - 1
//         sr_index = (asc >> 7) & 0x0F;
//         channels = (asc >> 3) & 0x0F;
//     } else {
//         // Fallbacks if ASC is missing
//         sr_index = get_adts_sr_index(par->sample_rate);
//         #if LIBAVCODEC_VERSION_INT >= AV_VERSION_INT(59, 37, 100)
//             channels = par->ch_layout.nb_channels;
//         #else
//             channels = par->channels;
//         #endif
//     }

//     if (channels == 0) channels = 1;

//     int frame_length = packet->size + 7;

//     uint8_t adts[7];
//     adts[0] = 0xFF;
//     adts[1] = 0xF1; 
//     adts[2] = ((profile & 0x03) << 6) | ((sr_index & 0x0F) << 2) | ((channels & 0x04) >> 2);
//     adts[3] = ((channels & 0x03) << 6) | ((frame_length & 0x1800) >> 11);
//     adts[4] = ((frame_length & 0x07F8) >> 3);
//     adts[5] = ((frame_length & 0x0007) << 5) | 0x1F;
//     adts[6] = 0xFC;

//     adtsPayload.resize(frame_length);
//     memcpy(adtsPayload.data(), adts, 7);
//     memcpy(adtsPayload.data() + 7, packet->data, packet->size);

//     return true; 
// }

/**
 * =========================================================
 * Ingestion
 * =========================================================
 */

// FrameMetadata VideoIngestion::makeFrameMetadataV(AVPacket* packet, bool isKey) {
//     FrameMetadata meta;

//     // Grab the absolute Wall-Clock time in milliseconds
//     meta.epochMs = utils::getCurrentEpochMSTime();

//     meta.magic = WrapMagicNumber;
//     meta.frameSize = packet->size;

//     int64_t pts = packet->pts;
//     int64_t dts = packet->dts;

//     // Delegate math to the synchronizer
//     normalizeTimestamps(pts, dts, videoStreamIndex);

//     meta.pts = pts;
//     meta.dts = dts;

//     meta.isKeyFrame = isKey ? 1 : 0;
//     meta.codecID = videoCodecID;
//     meta.mediaType = static_cast<uint8_t>(MediaType::VIDEO);

//     return meta;

// }

// FrameMetadata VideoIngestion::makeFrameMetadataA(AVPacket* packet) {
//     // Construct the metadata cleanly
//     FrameMetadata meta;
//     meta.epochMs = utils::getCurrentEpochMSTime();

//     meta.magic = WrapMagicNumber;
//     meta.frameSize = packet->size;

//     int64_t pts = packet->pts;
//     int64_t dts = packet->dts;

//     // Delegate math to the synchronizer
//     normalizeTimestamps(pts, dts, audioStreamIndex);

//     meta.pts = pts;
//     meta.dts = dts;

//     meta.isKeyFrame = 0;
//     meta.codecID = audioCodecID;
//     meta.mediaType = static_cast<uint8_t>(MediaType::AUDIO);
//     return meta;
// }

void VideoIngestion::ingestVideo(AVPacket* packet) {
    if (av_bsf_send_packet(bsfCtx, packet) == 0) {
        AVPacket* bsfPacket = av_packet_alloc();

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
                // Let the new TS Muxer handle the byte alignment and metadata
                if (tsMuxer) {
                    tsMuxer->muxVideoPacket(bsfPacket, isKey);
                }

                // Original raw packet still goes to the disk writer
                packetToDiskWriter(bsfPacket);

            } catch(...) {
                Log::error(camName + " [ingestVideo] Caught exception.");
            }

            av_packet_unref(bsfPacket);
        }
        av_packet_free(&bsfPacket);
    }
}

void VideoIngestion::ingestAudio(AVPacket* packet) {
    if (waitForKeyFrame) return;

    try {
        if (transcoder.isInitialized) {
            transcoder.process(packet, [this](AVPacket* cleanPkt) {
                // The Transcoder gives us a 48kHz AAC frame.
                // We DO NOT manually inject ADTS! FFmpeg's TS Muxer natively does it.
                if (tsMuxer) {
                    tsMuxer->muxAudioPacket(cleanPkt);
                }
            });
        } else {
            // Un-transcoded stream directly to muxer
            if (tsMuxer) {
                tsMuxer->muxAudioPacket(packet);
            }
        }

        // Original raw packet still goes to the disk writer
        packetToDiskWriter(packet);

    } catch(...) {
        Log::error(camName + " [ingestAudio] Caught exception writing audio.");
    }
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

    // Safely destroy the Muxer before freeing the SHM channel
    if (tsMuxer) {
        delete tsMuxer;
        tsMuxer = nullptr;
    }

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