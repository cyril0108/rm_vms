#include "TSMuxer.h"
#include "utils/Time.h"

TSMuxer::TSMuxer(std::shared_ptr<ISharedMemory> shm, int shmChannelID)
    : shm(shm), shmChannelID(shmChannelID) {}

TSMuxer::~TSMuxer() {
    if (outCtx) {
        av_write_trailer(outCtx); // Flush remaining bytes to SHM
        if (avioCtx) {
            av_freep(&avioCtx->buffer);
            avio_context_free(&avioCtx);
        }
        avformat_free_context(outCtx);
    }
}

bool TSMuxer::init(AVCodecParameters* vPar, AVRational vTb, 
                   AVCodecParameters* aPar, AVRational aTb) {

    inVideoTimeBase = vTb;
    inAudioTimeBase = aTb;

    // Allocate the output context for MPEG-TS
    avformat_alloc_output_context2(&outCtx, nullptr, "mpegts", nullptr);
    if (!outCtx) return false;

    // Setup the Custom AVIO Interceptor
    avioBuffer = (uint8_t*)av_malloc(avioBufferSize);
    avioCtx = avio_alloc_context(avioBuffer, avioBufferSize, 1, this, nullptr, shmWriteCallback, nullptr);
    outCtx->pb = avioCtx;
    outCtx->flags |= AVFMT_FLAG_CUSTOM_IO;

    // Create the Output Video Stream
    if (vPar) {
        AVStream* outStream = avformat_new_stream(outCtx, nullptr);
        avcodec_parameters_copy(outStream->codecpar, vPar);
        outStream->codecpar->codec_tag = 0; // Clear codec tag to let FFmpeg auto-assign
        outVideoStreamIndex = outStream->index;
    }

    // Create the Output Audio Stream
    if (aPar) {
        AVStream* outStream = avformat_new_stream(outCtx, nullptr);
        avcodec_parameters_copy(outStream->codecpar, aPar);
        outStream->codecpar->codec_tag = 0;
        outAudioStreamIndex = outStream->index;
    }

    // Write the MPEG-TS Header (Generates PAT and PMT automatically!)
    if (avformat_write_header(outCtx, nullptr) < 0) {
        Log::error("[TSMuxer] Failed to write TS header.");
        return false;
    }

    return true;
}

bool TSMuxer::muxVideoPacket(AVPacket* pkt, bool isKey) {
    if (!outCtx || outVideoStreamIndex < 0) return false;

    isKeyFrame = isKey;

    // // Increment our incoming frame tracker
    // totalPacketsReceived++;
    // // Log exactly when a packet enters the muxer from the camera
    // if( totalPacketsReceived < 6 || (totalPacketsReceived % 10000) < 6 ) {
    //     Log::info("[Muxer Diagnostic]["+shm->ChannelName()+"]["+std::to_string(shmChannelID)+"] INCOMING Video Packet #" + std::to_string(totalPacketsReceived) + 
    //               " | Size: " + std::to_string(pkt->size) + 
    //               " | PTS: " + std::to_string(pkt->pts));
    // }

    AVPacket* clone = av_packet_alloc();
    av_packet_ref(clone, pkt);

    // Force the stream index and rescale the clock
    clone->stream_index = outVideoStreamIndex;
    av_packet_rescale_ts(clone, inVideoTimeBase, outCtx->streams[outVideoStreamIndex]->time_base);

    // --- Delegate to the shared guard ---
    enforceMonotonicity(clone, lastVideoDTS);

    // 
    int ret = av_interleaved_write_frame(outCtx, clone);

    // Write immediately to bypass the interleaving cache
    // int ret = av_write_frame(outCtx, clone);

    // // Force FFmpeg to empty its 4KB buffer into SHM instantly
    // if (outCtx->pb) {
    //     avio_flush(outCtx->pb);
    // }

    av_packet_free(&clone);
    return ret >= 0;
}

bool TSMuxer::muxAudioPacket(AVPacket* pkt) {
    if (!outCtx || outAudioStreamIndex < 0) return false;

    AVPacket* clone = av_packet_alloc();
    av_packet_ref(clone, pkt);

    // Force the stream index and rescale the clock
    clone->stream_index = outAudioStreamIndex;
    av_packet_rescale_ts(clone, inAudioTimeBase, outCtx->streams[outAudioStreamIndex]->time_base);

    // --- Delegate to the shared guard ---
    enforceMonotonicity(clone, lastAudioDTS);

    //
    int ret = av_interleaved_write_frame(outCtx, clone);

    // // Write immediately
    // int ret = av_write_frame(outCtx, clone);

    // // // Force flush
    // if (outCtx->pb) {
    //     avio_flush(outCtx->pb);
    // }

    av_packet_free(&clone);
    return ret >= 0;
}


void TSMuxer::enforceMonotonicity(AVPacket* pkt, int64_t& lastDTSTracker) {
    if (pkt->dts != AV_NOPTS_VALUE) {
        if (lastDTSTracker != AV_NOPTS_VALUE && pkt->dts <= lastDTSTracker) {
            // Bump the timestamp forward by exactly 1 tick
            int64_t shift = lastDTSTracker - pkt->dts + 1;
            pkt->dts += shift;
            if (pkt->pts != AV_NOPTS_VALUE) {
                pkt->pts += shift; // Keep PTS/DTS distance mathematically equal
            }
        }
        lastDTSTracker = pkt->dts;
    }
}

// --- The Interceptor Callback ---
// FFmpeg automatically calls this when it has a complete TS chunk ready for the network
int TSMuxer::shmWriteCallback(void* opaque, uint8_t* buf, int buf_size) {
    TSMuxer* muxer = static_cast<TSMuxer*>(opaque);

    // Increment our outgoing callback trackers
    // muxer->totalCallbacksTriggered++;
    // muxer->totalBytesMuxed += buf_size;

    // if( muxer->totalCallbacksTriggered < 10 || (muxer->totalCallbacksTriggered % 10000) < 10 ) {
    //     // Log exactly when FFmpeg decides to output container data
    //     Log::info("[Muxer Diagnostic]["+muxer->shm->ChannelName()+"]["+std::to_string(muxer->shmChannelID)+"] ---> FLUSHING TO SHM Callback #" + std::to_string(muxer->totalCallbacksTriggered) +
    //               " | Block Size: " + std::to_string(buf_size) + " bytes" +
    //               " | Total Bytes Emitted So Far: " + std::to_string(muxer->totalBytesMuxed));
    // }

    // Create a lightweight metadata wrapper for Go
    FrameMetadata meta;
    meta.epochMs = utils::getCurrentEpochMSTime();
    meta.magic = WrapMagicNumber; // Assumes this is available via SharedMemory.h
    meta.frameSize = buf_size;
    meta.isKeyFrame = muxer->isKeyFrame ? 1 : 0;

    // We can assign a mediaType specifically for "TS Container Chunks" 
    // so Go knows to just pipe it directly to the HTTP socket.
    meta.mediaType = static_cast<uint8_t>(MediaType::VIDEO); // Or a dedicated TS type if you mapped one

    // Write the raw TS container bytes directly into Shared Memory
    if (muxer->shm->WriteFrame(muxer->shmChannelID, meta, buf) < 0) {
        Log::error("[TSMuxer] Failed to write TS chunk to SHM.");
    }

    return buf_size;
}