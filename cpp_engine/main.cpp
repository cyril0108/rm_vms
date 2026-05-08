#include <iostream>
#include <string>
#include <vector>
#include <map>
#include <unordered_map>

#include "Log.h"
#include "VideoIngestion.h"

#include "Command.h"
#include "SharedMemory.h"

const std::shared_ptr<ISharedMemory> SHM = ISharedMemory::CreateInstance();


int main(int argc, char* argv[]) {
    // Optimize I/O
    std::ios_base::sync_with_stdio(false);
    std::cin.tie(NULL);

    // Capture the root path from the first command line argument
    std::string rootPath = "";
    if (argc > 1) {
        rootPath = argv[1];
        Log::info("Worker started with root storage path: " + rootPath);
    }

    // 
    std::map<int, std::unique_ptr<VideoIngestion>> activeCameras;

    std::string line;
    while (std::getline(std::cin, line)) {
        if (line == "EXIT") break;

        Command cmd = parseCommand(line);

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
                        Log::send("{\"status\":\"shm\", \"channumber\":" + std::to_string(chnNum) + ", \"size\":" + std::to_string(bufferSize) + "}");
                    }
                } catch (...) {
                    Log::error("Error initializing SharedMemory.");
                }
            }

            if(cmd.Name == "START") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();
                    std::string url = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("id, url:" + idStr + " " + url );

                    // Respond to Go
                    Log::send("{\"status\":\"starting\", \"cam\":" + idStr + "}");

                    // Run logic
                    int camID = std::stoi(idStr);
                    activeCameras[camID] = std::make_unique<VideoIngestion>(SHM, camID, url, rootPath);

                } catch (...) {
                    Log::error("Error starting video ingestion.");
                }
            }

            if(cmd.Name == "STOP") {
                try {

                    std::string idStr = cmd.Args.front();
                    cmd.Args.pop();

                    Log::info("stop cam id:" + idStr);

                    // Remove VI
                    int camID = std::stoi(idStr);
                    auto it = activeCameras.find(camID);
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

        }

    }

    Log::info("Worker shutting down. Closing Shared Memory.");
    SHM->Close();

    return 0;
}
