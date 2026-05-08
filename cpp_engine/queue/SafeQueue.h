#pragma once

#include <queue>
#include <mutex>
#include <condition_variable>
#include <iostream>

#include "Log.h"

template <typename T>
class SafeQueue {
private:
    std::queue<T> queue;
    std::mutex mtx;
    std::condition_variable cv;
    size_t max_size; // Max items before we start dropping frames

public:
    explicit SafeQueue(size_t limit = 1000) : max_size(limit) {}

    // PUSH: Used by the Ingestion Thread
    // Returns false if the queue is full so the caller can free the memory.
    bool push(T item) {
        std::unique_lock<std::mutex> lock(mtx);
        
        // Safety: If queue is full (Disk is too slow), reject the packet.
        if (queue.size() >= max_size) {
            Log::error("WARNING: Disk Writer Queue full! Dropping frame.");
            return false;
        }

        queue.push(item);
        lock.unlock(); // Unlock before notifying to save CPU cycles
        cv.notify_one(); // Wake up the writer thread

        return true;
    }

    // POP: Used by the Writer Thread
    // This blocks (sleeps) if the queue is empty.
    T pop() {
        std::unique_lock<std::mutex> lock(mtx);
        
        // Wait until queue is NOT empty
        cv.wait(lock, [this] { return !queue.empty(); });

        T item = queue.front();
        queue.pop();
        return item;
    }

    // Utility to check current size if needed
    size_t size() {
        std::lock_guard<std::mutex> lock(mtx);
        return queue.size();
    }
};