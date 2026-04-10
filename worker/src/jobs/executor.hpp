#pragma once
#include "../api/client.hpp"
#include "state.hpp"
#include <atomic>
#include <iostream>
#include <latch>
#include <memory>
#include <thread>

namespace Jobs {
    class Executor {
    public:
        Executor(std::shared_ptr<std::atomic<bool>>& cancelCtx,
                 std::shared_ptr<std::atomic<uint8_t>>& progress,
                 std::shared_ptr<std::atomic<bool>>& jobActive,
                 Api::Client& client, std::latch& latch)
            : m_cancelCtx(cancelCtx), m_progress(progress),
              m_jobActive(jobActive), mr_client(client), mr_latch(latch) {};

        void Start();

    private:
        void requestJob();
        void executeThread();

        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::shared_ptr<std::atomic<uint8_t>> m_progress;
        std::shared_ptr<std::atomic<bool>> m_jobActive;
        std::atomic<State> m_state = State::NoJobActive;
        std::atomic<bool> m_cancelRequestThread;
        std::jthread m_workerThread;
        Api::Client& mr_client;
        std::latch& mr_latch;
    };
} // namespace Jobs