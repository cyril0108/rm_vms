#include <iostream>
#include <string>
#include <chrono>

// Estimates the storage footprint of an RTSP stream in Megabytes per minute
double EstimateStreamSizeMBPerMinute(const char* rtsp_url);