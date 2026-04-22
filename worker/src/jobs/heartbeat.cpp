#include "heartbeat.hpp"
#include <chrono>
#include <nlohmann/json.hpp>
#include <thread>

namespace Jobs {
    struct heartbeat {
        std::string id;
        NLOHMANN_DEFINE_TYPE_INTRUSIVE(heartbeat, id, jobProgress, jobActive)
    };

    void Heartbeat::Run() { m_heartbeatThread = std::jthread(&Heartbeat::heartbeatLoop, this); }

    void Heartbeat::heartbeatLoop() {
        using namespace std::chrono_literals;
        while (!m_cancelCtx->load()) {
            mr_client.Request("/workers/heartbeat", "POST", heartbeat{.id = m_workerID});
            std::this_thread::sleep_for(4500ms);
        }
        mr_latch.count_down();
    }
} // namespace Jobs
