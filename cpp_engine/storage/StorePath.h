#pragma once

#include <string>

struct AVPacket;

class StorePath {
private:
    std::string rootPath;
    std::string profile;

public:
    // Allow injecting the root path via constructor (e.g., from config.json later)
    StorePath(const std::string& root = "./recordings", const std::string& prof = "main");

    // Generates path: /rootPath/cam{int}/profile/YYYY/MM/DD/{timestampms}.mkv
    std::string For(int camID, AVPacket* packet);
};