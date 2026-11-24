#pragma once
#include "../api/client.hpp"
#include "../logging/logger.hpp"
#include <atomic>
#include <latch>
#include <memory>
#include <string_view>
#include <thread>

namespace Jobs {
    class Heartbeat {
    public:
        enum class ServerStatus {
            hibernation,
            active,
        };

    public:
        Heartbeat(const std::shared_ptr<std::atomic<bool>>& cancelCtx, const std::shared_ptr<Logging::Logger>& logger, const std::shared_ptr<Api::Client>& client, const std::string& workerID,
                  std::latch& latch)
            : m_cancelCtx(cancelCtx), m_logger(logger), m_workerID(workerID), m_client(client), mr_latch(latch) {
            m_serverStatus = std::make_shared<std::atomic<ServerStatus>>(ServerStatus::active);
        }

        void Start();

        const std::shared_ptr<std::atomic<ServerStatus>>& GetAtomicStatus() { return m_serverStatus; }

    private:
        void heartbeatLoop();

        std::shared_ptr<std::atomic<ServerStatus>> m_serverStatus;
        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::shared_ptr<Logging::Logger> m_logger;
        std::shared_ptr<Api::Client> m_client;
        std::string m_workerID;
        std::latch& mr_latch;
        std::jthread m_heartbeatThread;
    };
} // namespace Jobs
