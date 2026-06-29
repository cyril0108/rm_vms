#include <string>
#include "Log.h"

struct AVDictionary;

// Standard AVDictionary setups for NVR recording efficiency
AVDictionary* configureAVDictionary(AVDictionary* options);
// AVDictionary* setAVDictionaryAuth(AVDictionary* options, user, password);
void logUnusedOptions(AVDictionary* dict, const std::string& tag = "Stream");