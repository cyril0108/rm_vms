#include "EstimateStreamSize.h"

#include <thread>

extern "C" {
#include <libavformat/avformat.h>
}

#include "Log.h"

std::mutex cout_mutex;

// Helper function to perform a deep packet inspection when headers fail.
// It physically pulls packets for 'probe_duration_seconds' and weighs them.
double DeepProbeStreamSize(AVFormatContext* fmt_ctx, int probe_duration_seconds) {
    AVPacket* pkt = av_packet_alloc();
    if (!pkt) {
        std::cerr << "Error: Could not allocate AVPacket." << std::endl;
        return -1.0;
    }

    int64_t total_bytes_read = 0;
    auto start_time = std::chrono::steady_clock::now();
    bool first_packet_received = false;

    std::cout << "Initiating Deep Probe for " << probe_duration_seconds << " seconds..." << std::endl;

    // Pull raw packets from the RTSP stream
    while (av_read_frame(fmt_ctx, pkt) >= 0) {
        // Start the timer ONLY when the first packet arrives, 
        // avoiding network connection latency skewing our math.
        if (!first_packet_received) {
            start_time = std::chrono::steady_clock::now();
            first_packet_received = true;
        }

        total_bytes_read += pkt->size;
        av_packet_unref(pkt); // Instantly free the packet memory to prevent leaks

        // Check elapsed wall-clock time
        auto current_time = std::chrono::steady_clock::now();
        auto elapsed_ms = std::chrono::duration_cast<std::chrono::milliseconds>(current_time - start_time).count();

        if (elapsed_ms >= (probe_duration_seconds * 1000)) {
            break; // We have collected enough data
        }
    }
    av_packet_free(&pkt);

    if (total_bytes_read == 0) {
        return 0.0; // Stream is dead or empty
    }

    // Calculate Bytes per second based on our physical sample
    double bytes_per_second = static_cast<double>(total_bytes_read) / probe_duration_seconds;

    // Add 5% container/network overhead tax
    bytes_per_second = bytes_per_second * 1.05;

    // Convert to Megabytes per minute
    double bytes_per_minute = bytes_per_second * 60.0;
    double megabytes_per_minute = bytes_per_minute / (1024.0 * 1024.0);

    return megabytes_per_minute;
}


// Helper function to estimate stream size using Resolution and FPS (The Kush Gauge)
// This is an O(1) mathematical heuristic that avoids downloading physical packets.
double EstimateByResolutionAndFPS(AVFormatContext* fmt_ctx) {
    int width = 0;
    int height = 0;
    double fps = 0.0;
    bool is_h265 = false;

    // Find the video stream to extract parameters
    for (unsigned int i = 0; i < fmt_ctx->nb_streams; i++) {
        if (fmt_ctx->streams[i]->codecpar->codec_type == AVMEDIA_TYPE_VIDEO) {
            width = fmt_ctx->streams[i]->codecpar->width;
            height = fmt_ctx->streams[i]->codecpar->height;
            
            // Extract FPS. avg_frame_rate is preferred, r_frame_rate is the fallback
            AVRational frame_rate = fmt_ctx->streams[i]->avg_frame_rate;
            if (frame_rate.num == 0 || frame_rate.den == 0) {
                frame_rate = fmt_ctx->streams[i]->r_frame_rate;
            }
            if (frame_rate.den > 0) {
                fps = static_cast<double>(frame_rate.num) / frame_rate.den;
            }

            if (fmt_ctx->streams[i]->codecpar->codec_id == AV_CODEC_ID_HEVC) {
                is_h265 = true;
            }
            break;
        }
    }

    Log::info("width, height, fps:" + std::to_string(width) + " " + std::to_string(height) + " " + std::to_string(fps) );

    // If we can't extract the basic dimensions, the heuristic fails
    if (width == 0 || height == 0 || fps <= 0.0) {
        return 0.0; 
    }

    // The Kush Gauge: Bitrate = Width * Height * FPS * MotionFactor
    // Motion Factor 0.08 is standard for typical security camera footage (H.264)
    double bits_per_pixel = 0.08; 
    double estimated_bps = width * height * fps * bits_per_pixel;

    // H.265 (HEVC) is roughly 40% to 50% more efficient than H.264
    if (is_h265) {
        estimated_bps *= 0.55; 
    }

    // Add generic 64 Kbps for audio track
    estimated_bps += 64000.0;

    // Convert to Megabytes per minute with a 5% network/container overhead tax
    double bytes_per_second = (estimated_bps / 8.0) * 1.05;
    return (bytes_per_second * 60.0) / (1024.0 * 1024.0);
}

// Estimates the storage footprint of an RTSP stream in Megabytes per minute
double EstimateStreamSizeMBPerMinute(const char* rtsp_url) {
    AVFormatContext* fmt_ctx = nullptr;
    AVDictionary* options = nullptr;

    // NVR Best Practice: Force TCP for RTSP to prevent UDP packet loss
    // during the probing phase, which can cause avformat_find_stream_info to hang.
    av_dict_set(&options, "rtsp_transport", "tcp", 0);
    // Add a strict timeout (e.g., 5 seconds in microseconds) to prevent deadlocks
    av_dict_set(&options, "stimeout", "5000000", 0);

    // Open the network stream
    if (avformat_open_input(&fmt_ctx, rtsp_url, nullptr, &options) != 0) {
        std::cerr << "Error: Could not open RTSP stream." << std::endl;
        av_dict_free(&options);
        return -1.0;
    }

    // Read the stream headers to populate the codec parameters
    if (avformat_find_stream_info(fmt_ctx, nullptr) < 0) {
        std::cerr << "Error: Could not find stream information." << std::endl;
        avformat_close_input(&fmt_ctx);
        return -1.0;
    }

    int64_t total_bit_rate_bps = 0;

    // Approach A: Check headers first (Fastest O(1) method)
    if (fmt_ctx->bit_rate > 0) {
        total_bit_rate_bps = fmt_ctx->bit_rate;
    } else {
        for (unsigned int i = 0; i < fmt_ctx->nb_streams; i++) {
            if (fmt_ctx->streams[i]->codecpar->bit_rate > 0) {
                total_bit_rate_bps += fmt_ctx->streams[i]->codecpar->bit_rate;
            }
        }
    }

    Log::info("total_bit_rate_bps:" + std::to_string(total_bit_rate_bps));

    double final_mb_per_minute = 0.0;

    // Approach B: Headers failed. Try the Resolution/FPS Heuristic before Deep Probing.
    if (total_bit_rate_bps == 0) {
        std::cout << "Warning: Bitrate missing from headers. Attempting Resolution/FPS heuristic..." << std::endl;
        final_mb_per_minute = EstimateByResolutionAndFPS(fmt_ctx);

        // Approach C: If the heuristic fails (missing FPS/dimensions), execute Deep Probe.
        if (final_mb_per_minute == 0.0) {
            std::cout << "Warning: Heuristic failed. Falling back to Deep Probe." << std::endl;
            final_mb_per_minute = DeepProbeStreamSize(fmt_ctx, 3);
        }
    } else {
        // We have the bitrate, use standard math
        double bytes_per_second = static_cast<double>(total_bit_rate_bps) / 8.0;
        bytes_per_second *= 1.05; // 5% overhead
        final_mb_per_minute = (bytes_per_second * 60.0) / (1024.0 * 1024.0);
    }

    avformat_close_input(&fmt_ctx);
    av_dict_free(&options);

    return final_mb_per_minute;
}


void HandleProbeCommand(std::string camID, std::string profile, std::string url) {
    // Run the heavy FFmpeg probe
    double size = EstimateStreamSizeMBPerMinute(url.c_str());

    // Print the JSON back to Go (Ensure std::cout is thread-safe using a mutex!)
    std::lock_guard<std::mutex> lock(cout_mutex);

    Log::send("{\"status\":\"ess\", \"cam\":" + camID + ", \"profile\": \"" + profile + "\", \"estimated_mb\":" + std::to_string(size) + "}");
}
