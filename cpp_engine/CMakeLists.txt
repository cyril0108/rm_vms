cmake_minimum_required(VERSION 3.10)
project(nvr_worker)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

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

