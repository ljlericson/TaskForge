#include "heartbeat.hpp"
#include <chrono>
#include <nlohmann/json.hpp>
#include <thread>

namespace Jobs {
    struct heartbeat {
        std::string id;
        NLOHMANN_DEFINE_TYPE_INTRUSIVE(heartbeat, id)
    };

    void Heartbeat::Run() {
        m_heartbeatThread = std::jthread(&Heartbeat::heartbeatLoop, this);
    }

    void Heartbeat::heartbeatLoop() {
        using namespace std::chrono_literals;
        while (!m_cancelCtx->load()) {
            // slightly less than every 5 seconds to account for transmission
            // time and to stay ahead of 5 second cut off
            mr_client.Request("/workers/heartbeat",
                              heartbeat{.id = m_workerID.data()});
            std::this_thread::sleep_for(4500ms);
        }
    }
} // namespace Jobs
