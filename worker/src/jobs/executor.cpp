#include "executor.hpp"
#include <nlohmann/json.hpp>

namespace Jobs {
    void Executor::Start() {
        m_workerThread = std::jthread(&Executor::requestJob, this);
    }

    void Executor::requestJob() {
        while (!m_cancelCtx->load() && !m_cancelRequestThread.load()) {
            if (m_state.load() == State::NoJobActive) {
            }
        }
        if (m_cancelCtx->load())
            mr_latch.count_down();
    }
} // namespace Jobs