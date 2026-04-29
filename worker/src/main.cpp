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
    Logger::Infoln(std::format("Worker started with ID \"{}\"", workerID).c_str());
    std::latch latch(2);
    std::shared_ptr<std::atomic<bool>> cancelCtx = std::make_shared<std::atomic<bool>>(false);

    // main processess
    Api::Client client(cancelCtx, j["serverAddress"].get<std::string>(), workerID, j["privateKeyPath"].get<std::string>());
    Jobs::Heartbeat heartbeat(cancelCtx, client, workerID, latch);
    Jobs::Executor executor(cancelCtx, client, latch);

    client.RegisterWorker();
    heartbeat.Start();
    executor.Start();

    latch.wait();
}
