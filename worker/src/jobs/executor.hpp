#pragma once
#include "../api/client.hpp"
#include "state.hpp"
#include <atomic>
#include <iostream>
#include <latch>
#include <memory>
#include <thread>

namespace Jobs {
    struct Jar {
        std::string url;
        std::string mainClass;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(Jar, url, mainClass)
    };

    struct Resources {
        int executors;
        int coresPerExecutor;
        int memoryPerExecutorMB;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(Resources, executors, coresPerExecutor,
                                       memoryPerExecutorMB)
    };

    struct Data {
        std::vector<std::string> input;
        std::string output;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(Data, input, output)
    };

    struct JobSpec {
        std::string jobName;
        Jar jar;
        Resources resources;
        Data data;

        std::vector<std::string> arguments;
        std::map<std::string, std::string> environment;

        int timeoutSeconds;
        int priority;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(JobSpec, jobName, jar, resources, data,
                                       arguments, environment, timeoutSeconds,
                                       priority)
    };

    class Executor {
    public:
        Executor(std::shared_ptr<std::atomic<bool>>& cancelCtx,
                 std::shared_ptr<std::atomic<uint8_t>>& progress,
                 std::shared_ptr<std::atomic<bool>>& jobActive,
                 Api::Client& client, std::latch& latch)
            : m_cancelCtx(cancelCtx), m_progress(progress),
              m_jobActive(jobActive), mr_client(client), mr_latch(latch) {};

        void Start();

    private:
        void requestJob();
        void executeThread();

        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::shared_ptr<std::atomic<uint8_t>> m_progress;
        std::shared_ptr<std::atomic<bool>> m_jobActive;
        std::atomic<State> m_state = State::NoJobActive;
        std::atomic<bool> m_cancelRequestThread;
        JobSpec m_jobSpec;
        std::jthread m_workerThread;
        Api::Client& mr_client;
        std::latch& mr_latch;
    };
} // namespace Jobs