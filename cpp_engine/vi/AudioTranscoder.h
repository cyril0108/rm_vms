extern "C" {
#include <libavcodec/avcodec.h>
#include <libswresample/swresample.h>
#include <libavutil/audio_fifo.h>
}
#include <functional>

struct AVStream;

// --- THE TRANSCODE SANITIZER ---
struct AudioTranscoder {
    AVCodecContext* decCtx = nullptr;
    AVCodecContext* encCtx = nullptr;
    SwrContext* swrCtx = nullptr;
    AVAudioFifo* fifo = nullptr;
    int64_t outPts = 0;
    bool isInitialized = false;

    ~AudioTranscoder();
    bool init(AVStream* inStream);
    void process(AVPacket* inPkt, std::function<void(AVPacket*)> onEncoded);
};