#include "heartbeat.hpp"
#include "../logger/logger.hpp"
#include <chrono>
#include <nlohmann/json.hpp>
#include <thread>

namespace Jobs {
    struct heartbeat {
        std::string id;
        NLOHMANN_DEFINE_TYPE_INTRUSIVE(heartbeat, id)
    };

    void Heartbeat::Start() {
        Logger::Infoln("heartbeat thread starting");
        m_heartbeatThread = std::jthread(&Heartbeat::heartbeatLoop, this);
    }

    void Heartbeat::heartbeatLoop() {
        Logger::Infoln(std::format("cancelCtx = {}", m_cancelCtx->load()));
        using namespace std::chrono_literals;
        while (!m_cancelCtx->load()) {
            Api::Response res = mr_client.Request("/workers/heartbeat", "POST", heartbeat{.id = m_workerID});
            if (!res) {
                Logger::Errln(res.body);
            }
            std::this_thread::sleep_for(4500ms);
        }
        mr_latch.count_down();
    }
} // namespace Jobs
