#include "heartbeat.hpp"
#include <chrono>
#include <nlohmann/json.hpp>
#include <thread>

namespace Jobs {
    struct heartbeat {
        std::string id;
        uint8_t jobProgress;
        bool jobActive;

        /*
        {
            "id": "worker1",
            "jobProgress": 0,
            "jobActive": false
        }
        */

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(heartbeat, id, jobProgress, jobActive)
    };

    void Heartbeat::Run() {
        m_heartbeatThread = std::jthread(&Heartbeat::heartbeatLoop, this);
    }

    void Heartbeat::heartbeatLoop() {
        using namespace std::chrono_literals;
        while (!m_cancelCtx->load()) {
            // slightly less than every 5 seconds to account for transmission
            // time and to stay ahead of 5 second cut off
            mr_client.Request("/workers/heartbeat", "POST",
                              heartbeat{.id = m_workerID,
                                        .jobProgress = m_progress->load(),
                                        .jobActive = m_jobActive->load()});

            std::this_thread::sleep_for(4500ms);
        }
        mr_latch.count_down();
    }
} // namespace Jobs
