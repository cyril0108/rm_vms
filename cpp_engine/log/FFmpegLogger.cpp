#include "FFmpegLogger.h"
#include "Log.h" // Your existing logging class
#include <mutex>
#include <cstdarg>

extern "C" {
    #include <libavutil/log.h>
}

// Default prefix for the main thread
thread_local std::string currentThreadLogPrefix = "[System] ";
thread_local bool muteThreadLogs = false;
std::mutex ffmpegLogMutex;

void customFFmpegLogCallback(void* ptr, int level, const char* fmt, va_list vl) {

    // MUTE GUARD: If the thread flipped the switch, ignore the log entirely!
    if (muteThreadLogs) return;

    // Ignore debug/verbose spam
    if (level > av_log_get_level()) return;

    char message[2048];
    vsnprintf(message, sizeof(message), fmt, vl);

    std::string msgStr(message);
    
    // FFmpeg log strings usually end with '\n', which breaks single-line JSON loggers. Strip it.
    if (!msgStr.empty() && msgStr.back() == '\n') {
        msgStr.pop_back();
    }

    // Lock to prevent console garbling from 16 concurrent cameras
    std::lock_guard<std::mutex> lock(ffmpegLogMutex);

    // Map FFmpeg levels to your custom Log class
    if (level <= AV_LOG_ERROR || level <= AV_LOG_FATAL) {
        Log::error(currentThreadLogPrefix + msgStr);
    } else if (level <= AV_LOG_WARNING) {
        // Log::warning(currentThreadLogPrefix + msgStr); // If you have a warning method
        Log::info("[WARN] " + currentThreadLogPrefix + msgStr);
    } else {
        Log::info(currentThreadLogPrefix + msgStr);
    }
}

void setupCustomFFmpegLogging() {
    // Clamp the global FFmpeg noise level so it only emits warnings/errors
    av_log_set_level(AV_LOG_WARNING); 

    // Override the global callback
    av_log_set_callback(customFFmpegLogCallback);
}