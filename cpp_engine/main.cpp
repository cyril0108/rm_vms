#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <unordered_map>

#include "Log.h"
#include "VideoIngestion.h"

#include "Command.h"
#include "SharedMemory.h"

#include "EstimateStreamSize.h"

const std::shared_ptr<ISharedMemory> SHM = ISharedMemory::CreateInstance();

std::string cameraKey(std::string camID, std::string profile) {
    return camID + profile;
}

int main(int argc, char* argv[]) {

    std::cout << "Starting NVR Worker v" << NVR_VERSION 
                  << " (Commit: " << NVR_GIT_HASH 
                  << " | Built: " << NVR_BUILD_TIME << ")" << std::endl;

    // Optimize I/O
    std::ios_base::sync_with_stdio(false);
    std::cin.tie(NULL);

    // Capture the root path from the first command line argument
    std::string rootPath = "";
    if (argc > 1) {
        rootPath = argv[1];
        Log::info("Worker started with root storage path: " + rootPath);
    }

    std::map<std::string, std::unique_ptr<VideoIngestion>> activeCameras;

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line == "EXIT") break;

        Command cmd = parseCommand(line);

        Log::info("Got command " + line);

        if(cmd.Name != "") {

            if(cmd.Name == "WORKER") {
                // Initialize SharedMemory
                try {
                    std::string worker = cmd.Args.front();
                    std::string name = ringBufferNameFor(worker);
                    // Initial with some basic
                    // 10 MB per channel (10 * 1024 * 1024)
                    // size_t bufferSize = 10485760;
                    size_t bufferSize = 3145728; // 3mb
                    int chnNum = 10;
                    if(SHM->Create(name, chnNum, bufferSize)==false){
                        Log::error("Failed to create RingBuffer for:" + name);
                        Log::send("{\"status\":\"shmerr\", \"worker\":\"" + name + "}\"");
                    } else {
                        Log::send("{\"status\":\"shm_ready\", \"channumber\":" + std::to_string(chnNum) + ", \"size\":" + std::to_string(bufferSize) + "}");
                    }
                } catch (...) {
                    Log::error("Error initializing SharedMemory.");
                }
            }

            if(cmd.Name == "START") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string profile = cmd.Args.front();
                    cmd.Args.pop();
                    std::string url = cmd.Args.front();
                    cmd.Args.pop();

                    // Recording status
                    std::string recordStr = cmd.Args.front();
                    cmd.Args.pop();
                    bool isRecording = (recordStr == "true" || recordStr == "1");


                    std::string key = cameraKey(idStr, profile);

                    Log::info("id, profile, url:" + idStr + " " + profile + " " + url );
                    Log::info("id, recording:" + idStr + " " + recordStr + "->" + (isRecording ? "true" : "false") );

                    // Respond to Go
                    Log::send("{\"status\":\"starting\", \"cam\":" + idStr + "}");

                    // Run logic
                    int camID = std::stoi(idStr);
                    activeCameras[key] = std::make_unique<VideoIngestion>(VideoIngestionConfig{
                        SHM, camID, url, rootPath, profile, isRecording
                    });

                } catch (...) {
                    Log::error("Error starting video ingestion.");
                }
            }

            if(cmd.Name == "PROBE") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string profile = cmd.Args.front();
                    cmd.Args.pop();
                    std::string url = cmd.Args.front();
                    cmd.Args.pop();

                    double eSize = EstimateStreamSizeMBPerMinute(url.c_str());
                    // Respond to Go
                    Log::send("{\"status\":\"ess\", \"cam\":" + idStr + ", \"profile\": \"" + profile + "\", \"estimated_mb\":" + std::to_string(eSize) + "}");

                } catch (...) {
                    Log::error("Error probing stream size.");
                }
            }


            if(cmd.Name == "STOP") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string profile = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("stop cam id:" + idStr + " " + profile );

                    std::string key = cameraKey(idStr, profile);

                    // Remove VI
                    auto it = activeCameras.find(key);
                    if (it != activeCameras.end()) {

                        // Extract the unique_ptr from the map
                        std::unique_ptr<VideoIngestion> viToKill = std::move(it->second);

                        // Erase the empty map entry
                        activeCameras.erase(it);

                        std::thread([vi = std::move(viToKill)]() mutable {
                            // As soon as this lambda scope ends, 'vi' is destroyed, 
                            // triggering the destructor and joining the threads safely in the background.
                        }).detach();

                    }

                    // Respond to Go
                    // Log::send("{\"status\":\"stop\", \"cam\":" + idStr + "}");

                } catch (...) {
                    Log::error("Error stopping video ingestion.");
                }
            }

            // ===========================================
            // Recording Control
            // ===========================================
            if(cmd.Name == "RECORDING") {
                try {
                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string profile = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("start cam recording id:" + idStr);

                    std::string key = cameraKey(idStr, profile);

                    // Remove VI
                    auto it = activeCameras.find(key);
                    if (it != activeCameras.end()) {

                        it->second->startRecording();

                    }

                } catch (...) {
                    Log::error("Error starting recording.");
                }

            }

            if(cmd.Name == "NORECORDING") {
                try {
                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string profile = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("stop cam recording id:" + idStr);

                    std::string key = cameraKey(idStr, profile);

                    // Remove VI
                    auto it = activeCameras.find(key);
                    if (it != activeCameras.end()) {

                        it->second->stopRecording();

                    }


                } catch (...) {
                    Log::error("Error stopping recording.");
                }

            }


        }

    }

    Log::info("Worker shutting down. Closing Shared Memory.");
    SHM->Close();

    return 0;
}
