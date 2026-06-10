#include "SharedMemory.h"
#include <stack>
#include <vector>
#include <cstring>
#include <iostream>

// Linux/macOS specific headers
#if defined(__linux__) || defined(__APPLE__)
#include <sys/mman.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <unistd.h>

#include "Log.h"

std::string ringBufferNameFor(std::string worker) {
    return "shm"+worker;
}

class LinuxSharedMemory : public ISharedMemory {
private:
    std::string _name;
    int _shm_fd = -1;
    uint8_t* _basePtr = nullptr;
    size_t _totalSize = 0;
    int _maxChannel = 0;
    // int _lastChannelId = 0;

    // Helper struct to cache pointers for each channel
    // This prevents recalculating offsets on every single frame write
    struct ChannelCtx {
        RingBufferHeader* header;
        uint8_t* dataStart;
        uint32_t capacity;
    };
    std::vector<ChannelCtx> _channels;
    std::stack<int>         _availChannels;
    std::unordered_map<int, int> _channelForCam;
    std::mutex _channelMutex;

public:
    bool Create(const std::string& name, int numChannels, size_t sizePerChannel) override {
        // Setup Name (POSIX requires leading slash)
        _name = (name[0] == '/') ? name : "/" + name;

        _maxChannel = numChannels;

        // Calculate Total Size
        // Layout: [Header 0][Data 0] ... [Header N][Data N]
        size_t channelTotalSize = sizeof(RingBufferHeader) + sizePerChannel;
        _totalSize = channelTotalSize * numChannels;

        // Open Shared Memory
        _shm_fd = shm_open(_name.c_str(), O_CREAT | O_RDWR, 0666);
        if (_shm_fd == -1) {
            Log::PError("shm_open failed");
            return false;
        }

        // Resize
        if (ftruncate(_shm_fd, _totalSize) == -1) {
            Log::PError("ftruncate failed");
            close(_shm_fd);
            return false;
        }

        // Map
        void* map = mmap(0, _totalSize, PROT_READ | PROT_WRITE, MAP_SHARED, _shm_fd, 0);
        if (map == MAP_FAILED) {
            Log::PError("mmap failed");
            close(_shm_fd);
            return false;
        }
        _basePtr = static_cast<uint8_t*>(map);

        // Initialize Channel Contexts
        // We pre-calculate pointers so WriteFrame is fast
        uint8_t* cursor = _basePtr;
        _channels.reserve(numChannels);

        std::string strChanNum = std::to_string(numChannels);
        std::string strPerSize = std::to_string(sizePerChannel);

        // Log::info("[SHM]Created " + _name + "->" + " size:" + strPerSize);
        // Log::info("[SHM] total size: " + std::to_string(_totalSize));

        for(int i = 0; i < numChannels; ++i) {
            ChannelCtx ctx;
            ctx.header = reinterpret_cast<RingBufferHeader*>(cursor);

            // Move cursor past the header to finding the data start
            cursor += sizeof(RingBufferHeader);
            ctx.dataStart = cursor;
            ctx.capacity = static_cast<uint32_t>(sizePerChannel);

            // Initialize Header (Reset state)
            ctx.header->streamID = i;
            ctx.header->bufferCapacity = ctx.capacity;
            ctx.header->writeHead.store(0);
            ctx.header->readTail.store(0);

            _channels.push_back(ctx);
            _availChannels.push(i);

            // Move cursor to next channel
            cursor += sizePerChannel;
        }

        return true;
    }

    std::string ChannelName() {
        return _name;
    }

    void ReleaseChannelForCamID(int camID) {

        // Lock for thread safety
        std::lock_guard<std::mutex> lock(_channelMutex);

        auto it = _channelForCam.find(camID);

        if (it != _channelForCam.end()) {

            // Get the channel ID
            int chID = it->second;

            _channelForCam.erase(it);
            _availChannels.push(chID);
            CleanUpChannel(chID);

        }

    }

    void CleanUpChannel(int chID) {
        ChannelCtx& ch = _channels[chID];
        ch.header->writeHead.store(0, std::memory_order_relaxed);
        ch.header->readTail.store(0, std::memory_order_relaxed);
    }

    // Return available relased channel id, -1 when there is none.
    int PopAvailableChannel() {
        if(!_availChannels.empty()) {
            int val = _availChannels.top();
            _availChannels.pop();
            return val;
        }
        return -1;
    }

    int ChannelForCamID(int camID) {

        // Lock the mutex for thread safety
        std::lock_guard<std::mutex> lock(_channelMutex);

        auto it = _channelForCam.find(camID);
        if (it != _channelForCam.end()) {
            return it->second;
        }

        // Key does not exist, assign new one
        int newChId = PopAvailableChannel();
        if(newChId >= 0) {
            _channelForCam[camID] = newChId;
        } else {
            Log::error("[SHM] Failed to assign channel for Cam " + std::to_string(camID) + ". No channels available.");
        }

        // Return -1 when ther is no channel available
        return newChId;

    }

    // Returns false if buffer is full
    bool WriteFrame(int channelIdx, const FrameMetadata& meta, const uint8_t* data) override {
        if (channelIdx < 0 || channelIdx >= _channels.size()) return false;

        ChannelCtx& ch = _channels[channelIdx];

        // Load Atomic Indices
        // 'relaxed' is fine for load; we only need strict ordering when we commit
        uint32_t head = ch.header->writeHead.load(std::memory_order_relaxed);
        // readTail is handled by reader (Go Worker) for marking reading process.
        uint32_t tail = ch.header->readTail.load(std::memory_order_acquire);

        size_t size = static_cast<size_t>(meta.frameSize);
        size_t totalNeeded = sizeof(FrameMetadata) + size;

        // Calculate Next Position
        uint32_t nextHead = head + totalNeeded;
        bool wrapped = false;

        // Wrap-Around Check
        // If data doesn't fit at the end, we wrap to the beginning (Index 0)
        // We do NOT split frames across the boundary (simplifies Go reader)
        if (nextHead > ch.capacity) {
            nextHead = totalNeeded;
            wrapped = true;
        }

        // Collision Check (Head should not overtake Tail)
        // If we wrapped, we are writing at offset 0.
        uint32_t effectiveWriteStart = wrapped ? 0 : head;

        // Check if the area we want to write to overlaps with the "active" data area
        // Active Area = [Tail, Head)
        // Note: This logic assumes a non-full buffer. Full buffer detection is tricky lock-free.
        // Simplified check:
        if (effectiveWriteStart < tail && nextHead > tail) {
            // Buffer is full (Reader is too slow)
            return false;
        }

        uint8_t* writePtr = ch.dataStart + effectiveWriteStart;
        memcpy(writePtr, &meta, sizeof(FrameMetadata));

        // Write Video Data
        memcpy(writePtr + sizeof(FrameMetadata), data, size);

        // Commit (Atomic Store)
        // 'release' ensures the Go reader sees the data ONLY after the index is updated
        ch.header->writeHead.store(nextHead, std::memory_order_release);

        return true;
    }

    uint8_t* GetBuffer() const override {
        return _basePtr;
    }

    void Close() override {
        if (_basePtr) {
            munmap(_basePtr, _totalSize);
            _basePtr = nullptr;
        }
        if (_shm_fd != -1) {
            close(_shm_fd);
            _shm_fd = -1;
        }
    }
};

std::shared_ptr<ISharedMemory> ISharedMemory::CreateInstance() {
    return std::make_shared<LinuxSharedMemory>();
}

#endif