#include "SegmentRecorder.h"
#include <iostream>

#include "Log.h"

SegmentRecorder::~SegmentRecorder() {
    StopSegment();
}

bool SegmentRecorder::StartSegment(const std::string& filename, AVStream* inVideoStream, AVStream* inAudioStream) {
    if (isRecording) StopSegment();
    currentFilename = filename;

    Log::info("[SegmentRecorder] started for file \n " + currentFilename);

    // Reset DTS trackers for the new file timeline
    lastVideoDTS = AV_NOPTS_VALUE;
    lastAudioDTS = AV_NOPTS_VALUE;

    firstVideoPTS = AV_NOPTS_VALUE;
    lastVideoPTS = AV_NOPTS_VALUE;

    // Reset Timeline Normalization trackers
    hasStartTime = false;
    hasAudioStartTime = false;
    startVideoTime = AV_NOPTS_VALUE;
    startAudioTime = AV_NOPTS_VALUE;

    avformat_alloc_output_context2(&outFormatCtx, nullptr, "matroska", filename.c_str());
    if (!outFormatCtx) return false;

    // --- Setup Video Stream ---
    if (inVideoStream) {
        AVStream* outVStream = avformat_new_stream(outFormatCtx, nullptr);
        avcodec_parameters_copy(outVStream->codecpar, inVideoStream->codecpar);
        outVStream->codecpar->codec_tag = 0;

        outVideoStreamIndex = outVStream->index;
        inVideoStreamIndex = inVideoStream->index;
        videoInputTimeBase = inVideoStream->time_base;
    }

    // --- Setup Audio Stream ---
    if (inAudioStream) {

        // MKV is a universal container. We no longer need the MP4 strict compliance check.
        AVStream* outAStream = avformat_new_stream(outFormatCtx, nullptr);
        avcodec_parameters_copy(outAStream->codecpar, inAudioStream->codecpar);
        outAStream->codecpar->codec_tag = 0;

        outAudioStreamIndex = outAStream->index;
        inAudioStreamIndex = inAudioStream->index;
        audioInputTimeBase = inAudioStream->time_base;

    }

    // --- Open File & Write Header ---
    if (!(outFormatCtx->oformat->flags & AVFMT_NOFILE)) {
        if (avio_open(&outFormatCtx->pb, filename.c_str(), AVIO_FLAG_WRITE) < 0) {
            avformat_free_context(outFormatCtx);
            return false;
        }
    }

    AVDictionary* opts = nullptr;
    if (avformat_write_header(outFormatCtx, &opts) < 0) {
        return false;
    }
    av_dict_free(&opts);

    isRecording = true;
    return true;
}

void SegmentRecorder::WritePacket(AVPacket* packet) {
    if (!isRecording || !outFormatCtx) return;

    AVStream* outStream = nullptr;
    AVRational inputTimeBase;
    int64_t* lastDTS = nullptr;

    // Route the packet and grab the correct timebase
    if (packet->stream_index == inVideoStreamIndex) {

        outStream = outFormatCtx->streams[outVideoStreamIndex];
        inputTimeBase = videoInputTimeBase;
        packet->stream_index = outVideoStreamIndex; // Re-map to MP4 index
        lastDTS = &lastVideoDTS;

    } else if (inAudioStreamIndex != -1 && packet->stream_index == inAudioStreamIndex) {

        outStream = outFormatCtx->streams[outAudioStreamIndex];
        inputTimeBase = audioInputTimeBase;
        packet->stream_index = outAudioStreamIndex; // Re-map to MP4 index
        lastDTS = &lastAudioDTS;

    } else {
        // Unknown stream (e.g., metadata). Safely ignore.
        return; 
    }

    // ==========================================================
    // TIMELINE NORMALIZATION (The Zero-Offset Fix)
    // ==========================================================
    if (!normalizeTimeline(packet)) {
        return; // Packet was rejected (e.g., audio arrived before video anchor)
    }

    // ==========================================================
    // RESCALE TIMESTAMPS
    // ==========================================================
    av_packet_rescale_ts(packet, inputTimeBase, outStream->time_base);

    // ==========================================================
    // PTS/DTS Sanitization
    // ==========================================================
    sanitizeTimestamps(packet, lastDTS);

    if (packet->stream_index == outVideoStreamIndex) {
        if (firstVideoPTS == AV_NOPTS_VALUE) {
            firstVideoPTS = packet->pts;
        }
        lastVideoPTS = packet->pts;
    }

    // Interleave and write (FFmpeg handles the internal buffering to keep A/V in sync)
    if (av_interleaved_write_frame(outFormatCtx, packet) < 0) {
        Log::error("Error writing interleaved packet to file.");
    }
}

void SegmentRecorder::StopSegment() {
    if (isRecording && outFormatCtx) {
        av_write_trailer(outFormatCtx);
        if (!(outFormatCtx->oformat->flags & AVFMT_NOFILE)) {
            avio_closep(&outFormatCtx->pb);
        }
        avformat_free_context(outFormatCtx);
        outFormatCtx = nullptr;
        Log::info("[SegmentRecorder] closing segment \n " + currentFilename);
    }
    isRecording = false;
}


bool SegmentRecorder::normalizeTimeline(AVPacket* packet) {
    // Anchor the Video Clock
    if (!hasStartTime) {
        if (packet->stream_index == outVideoStreamIndex) {
            startVideoTime = (packet->dts == AV_NOPTS_VALUE) ? packet->pts : packet->dts;
            hasStartTime = true;
            Log::info("[SegmentRecorder] startVideoTime \n " + std::to_string(startVideoTime) + " : " + std::to_string(packet->pts));
        } else {
            // Drop any stray audio/metadata until the video timeline is anchored
            return false; 
        }
    }

    // Anchor the Audio Clock Independently
    if (!hasAudioStartTime && packet->stream_index == outAudioStreamIndex) {
        startAudioTime = (packet->dts == AV_NOPTS_VALUE) ? packet->pts : packet->dts;
        hasAudioStartTime = true;
        Log::info("[SegmentRecorder] startAudioTime \n " + std::to_string(startVideoTime) + " : " + std::to_string(packet->pts));
    }

    // Apply the offset and clamp negative values
    if (packet->stream_index == outVideoStreamIndex) {
        if (packet->pts != AV_NOPTS_VALUE) {
            packet->pts -= startVideoTime;
            if (packet->pts < 0) packet->pts = 0; 
        }
        if (packet->dts != AV_NOPTS_VALUE) {
            packet->dts -= startVideoTime;
            if (packet->dts < 0) packet->dts = 0;
        }
        // Log::info("[SegmentRecorder][normalizeTimeline]video dts " + std::to_string(packet->dts) + " : " + std::to_string(packet->pts));
    } else if (packet->stream_index == outAudioStreamIndex && hasAudioStartTime) {
        if (packet->pts != AV_NOPTS_VALUE) {
            packet->pts -= startAudioTime;
            if (packet->pts < 0) packet->pts = 0;
        }
        if (packet->dts != AV_NOPTS_VALUE) {
            packet->dts -= startAudioTime;
            if (packet->dts < 0) packet->dts = 0;
        }
        // Log::info("[SegmentRecorder][normalizeTimeline]audio dts " + std::to_string(packet->dts) + " : " + std::to_string(packet->pts));
    }


    return true;
}


void SegmentRecorder::sanitizeTimestamps(AVPacket* packet, int64_t* lastDTS) {
    if (*lastDTS != AV_NOPTS_VALUE && packet->dts <= *lastDTS) {
        // Calculate the exact offset needed to make this strictly increasing (+1)
        int64_t offset = *lastDTS - packet->dts + 1;

        packet->dts += offset;
        packet->pts += offset; 
    }

    // Update the tracker for the next packet
    *lastDTS = packet->dts;

    // Log::info("[SegmentRecorder][sanitizeTimestamps] lastDTS " + std::to_string(packet->dts) + " : " + std::to_string(packet->pts));

}

double SegmentRecorder::GetVideoDurationMilliseconds() const {
    if (firstVideoPTS == AV_NOPTS_VALUE || lastVideoPTS == AV_NOPTS_VALUE || !outFormatCtx) {
        return 0.0;
    }
    AVStream* outStream = outFormatCtx->streams[outVideoStreamIndex];
    int64_t duration = lastVideoPTS - firstVideoPTS;

    // Convert FFmpeg timebase to real seconds, then to milliseconds
    return duration * av_q2d(outStream->time_base) * 1000.0;
}