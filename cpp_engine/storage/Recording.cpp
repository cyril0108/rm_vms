#include "Recording.h"

#include "StorePath.h"
#include "Log.h"
#include "utils/Time.h"

#include <chrono>
#include <string>
#include <filesystem>

extern "C" {
#include <libavformat/avformat.h>
}

const int RECORDING_ROTATE_TIME = 1; // By minute



RecorderWorker::RecorderWorker(std::string rp, std::string prof) : rootPath(rp), profile(prof) {}

long RecorderWorker::getEndTimeMs(SegmentRecorder& recorder) {
    double duration = recorder.GetVideoDurationMilliseconds();
    return currentStartTimeMs + static_cast<long>(duration);
}

// --- IPC Helper Function ---
void RecorderWorker::sendSegmentDoneIPC(int camID, long startTimeMs, long endTimeMs, const std::string& filePath) {
    if (filePath.empty()) return;

    long sizeBytes = 0;
    std::error_code ec;
    sizeBytes = std::filesystem::file_size(filePath, ec);
    if (ec) sizeBytes = 0;
    // std::string profile = "main";

    // Construct and send JSON down the STDOUT pipe
    std::string json = "{\"status\":\"segment_done\", "
                       "\"cam\":" + std::to_string(camID) + ", "
                       "\"profile\": \"" + profile + "\", "
                       "\"start_time\":" + std::to_string(startTimeMs) + ", "
                       "\"end_time\":" + std::to_string(endTimeMs) + ", "
                       "\"file_path\":\"" + filePath + "\", "
                       "\"size_bytes\":" + std::to_string(sizeBytes) + "}";

    Log::send(json);
}

void RecorderWorker::writerWorker(SafeQueue<AVPacket*>& queue, AVStream* inVideoStream, AVStream* inAudioStream, int camID) {
    StorePath pathGenerator;

    if(!rootPath.empty()) {
        pathGenerator = StorePath(rootPath, profile);
    }

    auto lastSwitchTime = std::chrono::steady_clock::now();
    bool isFirstSegment = true;

    long endTimeMs;

    // The loop runs infinitely until the destructor pushes a nullptr
    while (true) {
        AVPacket* packet = queue.pop();

        // Graceful shutdown signal received from VideoIngestion teardown
        if (!packet) {
            break; 
        }

        if (packet->size == 0) {
            finalizeCurrentSegment(camID);
            
            isFirstSegment = true; 
            av_packet_unref(packet);
            av_packet_free(&packet);
            continue;
        }

        auto now = std::chrono::steady_clock::now();
        bool timeToRotate = std::chrono::duration_cast<std::chrono::minutes>(now - lastSwitchTime).count() >= RECORDING_ROTATE_TIME;
        // bool timeToRotate = std::chrono::duration_cast<std::chrono::seconds>(now - lastSwitchTime).count() >= 30;

        // CRITICAL A/V FIX: Ensure we only rotate files on a VIDEO Keyframe.
        // Audio streams often mark every packet as a keyframe.
        bool isVideoPacket = (inVideoStream && packet->stream_index == inVideoStream->index);
        bool isVideoKeyframe = isVideoPacket && (packet->flags & AV_PKT_FLAG_KEY);

        // Start the first file or rotate with Video Keyframe boundary
        if (isFirstSegment || (timeToRotate && isVideoKeyframe)) {

            finalizeCurrentSegment(camID);

            // Generate the precise directory tree and filename (e.g., /recordings/cam01/2026/03/12/15-30-00.mp4)
            currentFilePath = pathGenerator.For(camID, packet); 
            currentStartTimeMs = utils::getCurrentEpochMSTime();

            // Pass both streams to the SegmentRecorder so it can allocate the MP4 tracks
            recorder.StartSegment(currentFilePath, inVideoStream, inAudioStream);

            lastSwitchTime = now;
            isFirstSegment = false;

        }

        // Delegate A/V routing, rescaling, and interleaving to the recorder
        recorder.WritePacket(packet);

        // Free the cloned packet memory allocated by av_packet_ref
        av_packet_unref(packet);
        av_packet_free(&packet);
    }

    finalizeCurrentSegment(camID);

}

void RecorderWorker::finalizeCurrentSegment(int camID) {
    // Guard clause: Do nothing if we aren't actively recording
    if (!recorder.IsRecording()) {
        return;
    }

    long endTimeMs = getEndTimeMs(recorder);
    recorder.StopSegment();

    // Fire the IPC event for the completed file
    sendSegmentDoneIPC(camID, currentStartTimeMs, endTimeMs, currentFilePath);

    // CRITICAL: Clear the path so the next segment initialization 
    // or duplicate flush commands do not trigger an empty IPC message
    currentFilePath = ""; 
}