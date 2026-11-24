#include "../api/client.hpp"
#include "../jobs/executor.hpp"
#include "../jobs/heartbeat.hpp"
#include "../logging/logger.hpp"
#include <atomic>
#include <expected>
#include <latch>
#include <memory>
#include <optional>
#include <string>

namespace App {
    class Worker {
    public:
        void Run();

        static auto Create(const std::string& cfgFpath) -> std::expected<std::unique_ptr<Worker>, std::string>;

    private:
        Worker(const std::shared_ptr<std::atomic<bool>>& cancelCtx, const std::string& workerID, const std::shared_ptr<Logging::Logger>& logger, const std::shared_ptr<Api::Client>& client);

    private:
        std::shared_ptr<Logging::Logger> m_logger = nullptr;
        std::shared_ptr<Api::Client> m_client = nullptr;
        std::unique_ptr<Jobs::Executor> m_executor = nullptr;
        std::unique_ptr<Jobs::Heartbeat> m_heartbeat = nullptr;
        std::latch m_latch;
    };
} // namespace App