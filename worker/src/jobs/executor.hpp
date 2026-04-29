#pragma once
#include "../api/client.hpp"
#include "state.hpp"
#include <atomic>
#include <chrono>
#include <iostream>
#include <latch>
#include <memory>
#include <mutex>
#include <string_view>
#include <thread>

namespace Jobs {
    struct Jar {
        std::string url;
        std::string mainClass;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(Jar, url, mainClass)
    };

    struct Resources {
        int executionMemory;
        int executionCores;

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(Resources, executionCores, executionMemory)
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

        NLOHMANN_DEFINE_TYPE_INTRUSIVE(JobSpec, jobName, jar, resources, data, arguments, environment, timeoutSeconds, priority)
    };

    class Executor {
    public:
        Executor(std::shared_ptr<std::atomic<bool>>& cancelCtx, Api::Client& client, std::latch& latch) : m_cancelCtx(cancelCtx), mr_client(client), mr_latch(latch), m_execThreadLatch{1} {};

        void Start();

        void CancelActiveJob();

    private:
        void requestLoop();
        void executionLoop();

        void downloadFile(const std::string& url, const std::string& outputPath);
        void uploadFile(const std::string& filePath, const std::string& signedUploadURL);
        void runJob(const JobSpec& job);
        void writeRunScript(const JobSpec& job, const std::string& dir);
        std::vector<std::string> buildDockerArgs(const JobSpec& job, const std::string& dir);
        void runCommand(const std::vector<std::string>& args);
        int runProcessStreaming(const std::vector<std::string>& args, int timeoutSec);
        void parseProgress(const std::string& line);

        JobSpec m_jobSpec;
        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::string m_workspaceBase = "/tmp/taskforge";
        std::chrono::steady_clock::time_point m_jobStartTime;
        std::atomic<State> m_lastReportedState{State::NoJobActive};
        std::atomic<State> m_state = State::NoJobActive;
        std::atomic<pid_t> m_activePid{-1};
        std::atomic<bool> m_cancelJob{false};
        std::atomic<int> m_progress{0};
        std::atomic<int> m_phase{0};
        std::mutex m_jobMutex;
        std::jthread m_execThread;
        std::jthread m_requestThread;
        Api::Client& mr_client;
        std::latch& mr_latch;
        std::latch m_execThreadLatch;
    };
} // namespace Jobs