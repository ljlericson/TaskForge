#include "app.hpp"
#include <iostream>

namespace App {
    auto Worker::Create(const std::string& cfgFpath) -> std::expected<std::unique_ptr<Worker>, std::string> {
        std::ifstream file(cfgFpath);
        if (!file.is_open()) {
            return std::unexpected("config/worker.json open err");
        }
        nlohmann::json j;
        try {
            file >> j;
        } catch (const nlohmann::json::exception& e) {
            return std::unexpected(fmt::format("config/worker.json error: {}", e.what()));
        }

        if (!j.contains("logPath")) {
            return std::unexpected("missing \"logPath\" field in config");
        }
        if (!j.contains("id")) {
            return std::unexpected("missing \"id\" field in config");
        }
        if (!j.contains("serverAddress")) {
            return std::unexpected("missing \"serverAddress\" field in config");
        }
        if (!j.contains("privateKeyPath")) {
            return std::unexpected("missing \"privateKeyPath\" field in config");
        }

        // initialisations that could result in an error
        auto loggerResult = Logging::Logger::Create(j["logPath"]);
        if (!loggerResult) {
            return std::unexpected(loggerResult.error());
        }
        std::shared_ptr logger = loggerResult.value();

        std::shared_ptr<std::atomic<bool>> cancelCtx = std::make_shared<std::atomic<bool>>(false);
        auto clientResult = Api::Client::Create(cancelCtx, logger, j["serverAddress"].get<std::string>(), j["id"], j["privateKeyPath"].get<std::string>());
        if (!clientResult) {
            return std::unexpected(clientResult.error());
        }
        auto client = clientResult.value();

        return std::unique_ptr<Worker>(new Worker(cancelCtx, j["id"], logger, client));
    }

    Worker::Worker(const std::shared_ptr<std::atomic<bool>>& cancelCtx, const std::string& workerID, const std::shared_ptr<Logging::Logger>& logger, const std::shared_ptr<Api::Client>& client)
        : m_latch(2) {
        m_logger = logger;
        m_client = client;
        m_heartbeat = std::make_unique<Jobs::Heartbeat>(cancelCtx, m_logger, m_client, workerID, m_latch);
        m_executor = std::make_unique<Jobs::Executor>(cancelCtx, m_heartbeat->GetAtomicStatus(), m_logger, m_client, m_latch);
    }

    void Worker::Run() {
        auto res = m_client->RegisterWorker();
        if (!res || res.statusCode == 400) {
            m_logger->Errln(fmt::format("registration failed, server gave error: {}", res.body).c_str());
        } else {
            m_logger->Infoln("worker registered succsesfully");

            m_heartbeat->Start();
            m_executor->Start();

            m_latch.wait();
        }
    }

} // namespace App