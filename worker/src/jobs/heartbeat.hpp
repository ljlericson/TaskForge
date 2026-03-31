#pragma once
#include "../api/client.hpp"
#include <atomic>
#include <memory>
#include <string_view>
#include <thread>

namespace Jobs {
    class Heartbeat {
    public:
        Heartbeat(std::shared_ptr<std::atomic<bool>>& cancelCtx,
                  Api::Client& client, std::string_view workerID)
            : mr_client(client), m_cancelCtx(cancelCtx), m_workerID(workerID) {}

        void Run();

    private:
        void heartbeatLoop();

        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::shared_ptr<std::atomic<uint8_t>> m_progress;
        std::string_view m_workerID;
        Api::Client& mr_client;
        std::jthread m_heartbeatThread;
    };
} // namespace Jobs
