#pragma once
#include <string>
#include <memory>
#include <cstdint>
#include <atomic>
#include <unordered_map>

// Align to 64 bytes to prevent "False Sharing" on CPU cache lines
#define CACHE_LINE 64

#define WrapMagicNumber 0xFFAABBCC

enum class MediaType : uint8_t {
    VIDEO = 0,
    AUDIO = 1
};

struct FrameMetadata {
    uint32_t magic;        // 4 bytes (Offset 0)
    uint32_t frameSize;    // 4 bytes (Offset 4)  -> Sum: 8 bytes (Perfectly aligned for the next 8-byte var)
    
    uint64_t epochMs;      // 8 bytes (Offset 8)  -> Wall-clock time
    int64_t  pts;          // 8 bytes (Offset 16) -> Presentation Time
    int64_t  dts;          // 8 bytes (Offset 24) -> Decoding Time
    
    uint32_t codecID;      // 4 bytes (Offset 32)
    uint8_t  isKeyFrame;   // 1 byte  (Offset 36)
    uint8_t  mediaType;    // 1 byte  (Offset 37)
    uint8_t  _padding[26]; // 26 bytes(Offset 38) -> Total: EXACTLY 64 bytes
};

struct RingBufferHeader {
    std::atomic<uint32_t> writeHead;
    std::atomic<uint32_t> readTail;
    uint32_t bufferCapacity; // Renamed from bufferSize for clarity
    uint32_t streamID;
    uint8_t  _padding[48];
};

// Calculation:
// Buffer Base Address + sizeof(RingBufferHeader) = Start of Data

std::string ringBufferNameFor(std::string worker);

// ISharedMemory Interface
class ISharedMemory {
private:
    std::unordered_map<int, int> _channelForCam;

public:
    virtual ~ISharedMemory() = default;

    // Initialize Shared Memory for N channels
    // We replace the generic 'size' with specific channel config
    virtual bool Create(const std::string& name, int numChannels, size_t sizePerChannel) = 0;

    virtual int ChannelForCamID(int camID) = 0;
    virtual void ReleaseChannelForCamID(int camID);

    // Write a video frame to a specific channel (Thread-Safe via atomics)
    virtual bool WriteFrame(int channelIdx, const FrameMetadata& meta, const uint8_t* payload) = 0;

    // Get the raw pointer (mostly for debugging or manual inspection)
    virtual uint8_t* GetBuffer() const = 0;

    virtual void Close() = 0;

    // Static Factory Method: Returns the correct OS version
    static std::shared_ptr<ISharedMemory> CreateInstance();
};