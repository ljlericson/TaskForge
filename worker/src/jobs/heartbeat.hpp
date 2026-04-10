#pragma once
#include "../api/client.hpp"
#include <atomic>
#include <latch>
#include <memory>
#include <string_view>
#include <thread>

namespace Jobs {
    class Heartbeat {
    public:
        Heartbeat(std::shared_ptr<std::atomic<bool>>& cancelCtx,
                  std::shared_ptr<std::atomic<uint8_t>>& progress,
                  std::shared_ptr<std::atomic<bool>>& jobActive,
                  Api::Client& client, const std::string& workerID,
                  std::latch& latch)
            : m_cancelCtx(cancelCtx), m_progress(progress),
              m_jobActive(jobActive), m_workerID(workerID), mr_client(client),
              mr_latch(latch) {}

        void Run();

    private:
        void heartbeatLoop();

        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::shared_ptr<std::atomic<uint8_t>> m_progress;
        std::shared_ptr<std::atomic<bool>> m_jobActive;
        std::string m_workerID;
        Api::Client& mr_client;
        std::latch& mr_latch;
        std::jthread m_heartbeatThread;
    };
} // namespace Jobs
