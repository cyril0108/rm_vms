#include "StorePath.h"
#include <iostream>
#include <sstream>
#include <iomanip>
#include <chrono>
#include <filesystem> // Requires C++17
#include <ctime>

extern "C" {
#include <libavformat/avformat.h>
}

StorePath::StorePath(const std::string& root, const std::string& prof) : rootPath(root), profile(prof) {}

std::string StorePath::For(int camID, AVPacket* packet) {

    // Capture the exact wall-clock time the segment starts
    auto now = std::chrono::system_clock::now();
    
    // NEW: Extract 13-digit millisecond epoch for the filename
    auto ms_epoch = std::chrono::duration_cast<std::chrono::milliseconds>(now.time_since_epoch()).count();

    // Convert to time_t (seconds) strictly for local calendar math (folder structure)
    auto in_time_t = std::chrono::system_clock::to_time_t(now);

    std::tm bt{};

    // POSIX thread-safe localtime. Perfect for macOS and Linux.
    localtime_r(&in_time_t, &bt);

    // Construct the directory path: ROOT/camID/profile/YYYY/MM/DD
    std::ostringstream folderStream;
    folderStream << rootPath
                 << "/cam" << std::setfill('0') << std::setw(2) << camID
                 << "/" << profile
                 << "/" << std::put_time(&bt, "%Y/%m/%d");

    std::string folderPath = folderStream.str();

    // Ensure the directory structure exists (equivalent to `mkdir -p`)
    std::error_code ec;
    std::filesystem::create_directories(folderPath, ec);
    if (ec) {
        std::cerr << "[StorePath] Critical IO Error: Failed to create directories: " 
                  << ec.message() << std::endl;
    }

    // Construct the final filename using the millisecond epoch: 1716301234000.mkv
    std::ostringstream fileStream;
    fileStream << folderPath << "/" << ms_epoch << ".mkv";

    return fileStream.str();
}