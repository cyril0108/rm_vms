cmake_minimum_required(VERSION 3.10)

if(NOT DEFINED THE_VERSION OR THE_VERSION STREQUAL "")
    set(THE_VERSION "0.0.0")
endif()

project(nvr_worker VERSION ${THE_VERSION})

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)


# ==============================================================================
# VERSIONING METADATA EXTRACTION
# ==============================================================================
# Fetch the Git Hash if it does not exist
if(NOT DEFINED GIT_HASH OR GIT_HASH STREQUAL "unknown" OR GIT_HASH STREQUAL "")
    execute_process(
        COMMAND git rev-parse --short HEAD
        WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
        OUTPUT_VARIABLE GIT_HASH
        OUTPUT_STRIP_TRAILING_WHITESPACE
        ERROR_QUIET
    )
endif()

# Fallback just in case you compile outside of Git
if(NOT GIT_HASH)
    set(GIT_HASH "unknown")
endif()

# Fetch the current Timestamp
string(TIMESTAMP BUILD_TIME "%Y-%m-%dT%H:%M:%S")
# ==============================================================================


# Find FFmpeg Libraries
find_package(PkgConfig REQUIRED)
pkg_check_modules(AVCODEC REQUIRED libavcodec)
pkg_check_modules(AVFORMAT REQUIRED libavformat)
pkg_check_modules(AVUTIL REQUIRED libavutil)
# If you are using pkg-config, it looks like this:
pkg_check_modules(FFMPEG REQUIRED libavcodec libavformat libavutil libswresample)

include_directories(${AVCODEC_INCLUDE_DIRS} ${AVFORMAT_INCLUDE_DIRS} ${AVUTIL_INCLUDE_DIRS})
link_directories(${AVCODEC_LIBRARY_DIRS} ${AVFORMAT_LIBRARY_DIRS} ${AVUTIL_LIBRARY_DIRS})

# Set cpp compile list
file(GLOB_RECURSE VI_CPP "vi/*.cpp")
file(GLOB_RECURSE LOG_CPP "log/*.cpp")
file(GLOB_RECURSE MEMORY_CPP "memory/*.cpp")
file(GLOB_RECURSE UTILS_CPP "utils/*.cpp")
file(GLOB_RECURSE STORAGE_CPP "storage/*.cpp")
file(GLOB_RECURSE QUEUE_CPP "queue/*.cpp")
file(GLOB_RECURSE MUXER_CPP "muxer/*.cpp")

add_executable(nvr_worker main.cpp ${VI_CPP} ${LOG_CPP} ${UTILS_CPP} ${MEMORY_CPP} ${STORAGE_CPP} ${QUEUE_CPP} ${MUXER_CPP})

# Header search folders
target_include_directories(nvr_worker PRIVATE ${CMAKE_CURRENT_SOURCE_DIR}
                        ${CMAKE_CURRENT_SOURCE_DIR}/vi
                        ${CMAKE_CURRENT_SOURCE_DIR}/log
                        ${CMAKE_CURRENT_SOURCE_DIR}/utils
                        ${CMAKE_CURRENT_SOURCE_DIR}/memory
                        ${CMAKE_CURRENT_SOURCE_DIR}/storage
                        ${CMAKE_CURRENT_SOURCE_DIR}/queue
                        ${CMAKE_CURRENT_SOURCE_DIR}/muxer
                        )

target_link_libraries(nvr_worker
    ${AVCODEC_LIBRARIES}
    ${AVFORMAT_LIBRARIES}
    ${AVUTIL_LIBRARIES}
    pthread
    swresample
)


# ==============================================================================
# INJECT MACROS INTO C++
# ==============================================================================
target_compile_definitions(nvr_worker PRIVATE 
    NVR_VERSION="${PROJECT_VERSION}"
    NVR_GIT_HASH="${GIT_HASH}"
    NVR_BUILD_TIME="${BUILD_TIME}"
)