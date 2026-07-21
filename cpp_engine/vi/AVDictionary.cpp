#include "AVDictionary.h"

extern "C" {
#include <libavformat/avformat.h>
}

AVDictionary* configureTSMuxerAVDictionary(AVDictionary* options) {

    av_dict_set(&options, "flush_pes", "1", 0); // Flush PES packets immediately
    av_dict_set(&options, "mpegts_flags", "resend_headers", 0); // Helps clients lock on faster

    return options;

}

// Standard AVDictionary setups for NVR recording efficiency
AVDictionary* configureAVDictionary(AVDictionary* options) {

    // --- NETWORK STABILITY ---

    // av_dict_set(&options, "strict", "experimental", 0);

    // Force TCP. 
    // UDP (default) causes gray artifacts when packets drop. 
    // TCP guarantees every byte arrives, essential for evidence.
    av_dict_set(&options, "rtsp_transport", "tcp", 0);

    // Socket Timeout (in microseconds).
    // Default is usually "wait forever". 
    // Set to 5 seconds. If no data comes in 5s, kill the connection and reconnect.
    // This prevents "zombie" threads hanging your system.
    av_dict_set(&options, "stimeout", "5000000", 0); 

    // I/O Timeout
    // Acts as a fallback timeout for the underlying TCP connection phase 
    // (during avformat_open_input). Also in microseconds.
    av_dict_set(&options, "rw_timeout", "5000000", 0);

    // Minimize Latency
    // Increase the Kernel Socket Buffer.
    // 64 cameras x 8Mbps = Huge throughput.
    // If the buffer is too small, the OS drops packets before your app sees them.
    av_dict_set(&options, "buffer_size", "2097152", 0); // 2MB per socket


    // --- STARTUP SPEED & MEMORY (Crucial for 64 Cams) ---

    // Probe Size (in bytes).
    // Default is ~5MB. FFmpeg reads this much data just to guess the format.
    // 64 cams * 5MB = 320MB of RAM spiked just to connect.
    // Reduce to 32KB (usually enough for H.264 RTSP headers).
    // av_dict_set(&options, "probesize", "64768", 0); 
    av_dict_set(&options, "probesize", "5000000", 0);      // 5 MB (default is 50MB)

    // Analyze Duration (in microseconds).
    // How long to watch the stream to detect frame rate/resolution.
    // Default is 5 seconds. Reduce to 0.5 or 1 second.
    av_dict_set(&options, "analyzeduration", "500000", 0); // 0.5 seconds (default is 5,000,000)


    // ---  LATENCY REDUCTION ---

    // Flags to reduce internal buffering.
    // 'nobuffer': push packets immediately, don't aggregate them.
    // 'discardcorrupt': if a packet is broken, throw it away (don't try to fix it).
    av_dict_set(&options, "fflags", "nobuffer+discardcorrupt", 0);

    // Low delay flag for the codec context (if you were decoding, but safe to set generally)
    av_dict_set(&options, "flags", "low_delay", 0);

    av_dict_set(&options, "user_agent", "NVR_System", 0);

    return options;

}

// AVDictionary* setAVDictionaryAuth(AVDictionary* options, user, password) {

//     // av_dict_set(&options, "allowed_media_types", "video", 0); 

//     av_dict_set(&options, "default_protocol_opts", 
//             ("user=" + user + "&password=" + password).c_str(), 0);

//     return options;

// }


/**
 * Checks for any options that FFmpeg did NOT consume.
 * Useful for debugging typos (e.g. "timeout" vs "stimeout").
 * * @param dict The dictionary to check.
 * @param tag  Optional string to identify which stream this belongs to.
 */
void logUnusedOptions(AVDictionary* dict, const std::string& tag) {
    AVDictionaryEntry* t = nullptr;

    // Iterate over all remaining entries in the dictionary
    while ((t = av_dict_get(dict, "", t, AV_DICT_IGNORE_SUFFIX))) {
        Log::error("[Warning][" + tag + "] Unused Option: Key='" + t->key + "', Value='" + t->value + "'");
    }
}
