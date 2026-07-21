#pragma once
#include <string>

// thread_local ensures every std::thread gets its own independent copy of this variable
extern thread_local std::string currentThreadLogPrefix;
extern thread_local bool muteThreadLogs;

// Call this ONCE when the C++ worker boots up
void setupCustomFFmpegLogging();