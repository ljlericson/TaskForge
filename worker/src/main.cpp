#include "api/client.hpp"
#include "auth/signer.hpp"
#include "jobs/executor.hpp"
#include "jobs/heartbeat.hpp"
#include "logger/logger.hpp"
#include <atomic>
#include <chrono>
#include <format>
#include <fstream>
#include <latch>
#include <memory>
#include <nlohmann/json.hpp>
#include <thread>

int main() {
    std::ifstream file("config/worker.json");
    if (!file.is_open()) {
        Logger::Errln("config/worker.json open err");
        return -1;
    }
    nlohmann::json j;
    try {
        file >> j;
    } catch (const nlohmann::json::exception& e) {
        Logger::Errln(std::format("config/worker.json error: {}", e.what()));
        return -1;
    }

    // shared reasources
    std::string workerID = j["id"].get<std::string>();
    Logger::Infoln(
        std::format("Registering with worker ID \"{}\"", workerID).c_str());
    std::latch latch(2);
    std::shared_ptr<std::atomic<uint8_t>> jobProgress =
        std::make_shared<std::atomic<uint8_t>>(0);
    std::shared_ptr<std::atomic<bool>> jobActive =
        std::make_shared<std::atomic<bool>>(false);
    std::shared_ptr<std::atomic<bool>> cancelCtx =
        std::make_shared<std::atomic<bool>>(false);

    std::unique_ptr<Api::Client> client = std::make_unique<Api::Client>(
        cancelCtx, j["serverAddress"].get<std::string>(), workerID,
        j["privateKeyPath"].get<std::string>());

    std::unique_ptr<Jobs::Heartbeat> heartbeat =
        std::make_unique<Jobs::Heartbeat>(cancelCtx, jobProgress, jobActive,
                                          *client, workerID, latch);

    std::unique_ptr<Jobs::Executor> executor = std::make_unique<Jobs::Executor>(
        cancelCtx, jobProgress, jobActive, *client, latch);

    client->RegisterWorker();
    heartbeat->Run();
    executor->Start();

    latch.wait();
}
