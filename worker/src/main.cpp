#include "api/client.hpp"
#include "auth/signer.hpp"
#include "jobs/executor.hpp"
#include "jobs/heartbeat.hpp"
#include "logger/logger.hpp"
#include <atomic>
#include <format>
#include <fstream>
#include <memory>
#include <nlohmann/json.hpp>

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

    std::shared_ptr<std::atomic<bool>> cancelCtx =
        std::make_shared<std::atomic<bool>>(false);
    std::unique_ptr<Api::Client> client = std::make_unique<Api::Client>(
        j["serverAddress"].get<std::string_view>(), j["id"].get<std::string>(),
        j["privateKeyPath"].get<std::string>());
    std::unique_ptr<Jobs::Heartbeat> heartbeat =
        std::make_unique<Jobs::Heartbeat>(cancelCtx, *client,
                                          j["id"].get<std::string_view>());

    client->RegisterWorker();
    heartbeat->Run();

    while (true) {
        using namespace std::chrono_literals;
        std::this_thread::sleep_for(100ms);
    }
}
