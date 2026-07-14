#include <string>
#include "Log.h"

struct AVDictionary;

// Standard AVDictionary setups for NVR recording efficiency
AVDictionary* configureAVDictionary(AVDictionary* options);

// Standard AVDictionary setups for TSMuxer
AVDictionary* configureTSMuxerAVDictionary(AVDictionary* options);

// AVDictionary* setAVDictionaryAuth(AVDictionary* options, user, password);
void logUnusedOptions(AVDictionary* dict, const std::string& tag = "Stream");