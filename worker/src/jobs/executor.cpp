#include "executor.hpp"
#include <chrono>
#include <map>
#include <nlohmann/json.hpp>
#include <string>
#include <thread>
#include <vector>

using json = nlohmann::json;

struct null {
    std::string ID;
    NLOHMANN_DEFINE_TYPE_INTRUSIVE(null, ID)
};

namespace Jobs {
    void Executor::Start() {
        m_workerThread = std::jthread(&Executor::requestJob, this);
    }

    void Executor::requestJob() {
        while (!m_cancelCtx->load() && !m_cancelRequestThread.load()) {
            if (m_state.load() == State::NoJobActive) {
                using namespace std::chrono_literals;
                Api::Response response =
                    mr_client.Request("/jobs/next", "GET", null{});
                if (response.statusCode != 204) {
                    json j = json::parse(response.body);
                    m_jobSpec = j.get<JobSpec>();
                    m_state = State::JobActive;
                }
                std::this_thread::sleep_for(1s);
            }
        }
        if (m_cancelCtx->load())
            mr_latch.count_down();
    }
} // namespace Jobs