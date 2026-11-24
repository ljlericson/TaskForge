#include "heartbeat.hpp"
#include <chrono>
#include <fmt/format.h>
#include <nlohmann/json.hpp>
#include <thread>

namespace Jobs {
    struct heartbeat {
        std::string id;
        NLOHMANN_DEFINE_TYPE_INTRUSIVE(heartbeat, id)
    };

    void Heartbeat::Start() { m_heartbeatThread = std::jthread(&Heartbeat::heartbeatLoop, this); }

    void Heartbeat::heartbeatLoop() {
        uint8_t numOfFails = 0;
        using namespace std::chrono_literals;
        while (!m_cancelCtx->load()) {
            switch (m_serverStatus->load()) {
                case ServerStatus::active: {
                    Api::Response res = m_client->Request("/workers/heartbeat", "POST", heartbeat{.id = m_workerID});
                    if (!res) {
                        ++numOfFails;
                        m_logger->Errln(fmt::format("heartbeat failed (err: {}), number of missed heartbeats: {}", res.body, numOfFails));
                    }

                    if (numOfFails > 5) {
                        m_logger->Infoln("server is unresponsive, entering hibernation mode");
                        m_serverStatus->store(ServerStatus::hibernation);
                        numOfFails = 0;
                    }
                    std::this_thread::sleep_for(4500ms);
                    break;
                }
                case ServerStatus::hibernation: {
                    Api::Response res = m_client->RegisterWorker();
                    if (res) {
                        m_serverStatus->store(ServerStatus::active);
                        m_logger->Infoln("connection to server established");
                    } else {
                        std::this_thread::sleep_for(60s);
                    }
                }
            }
        }
        mr_latch.count_down();
    }
} // namespace Jobs
