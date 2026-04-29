#include "executor.hpp"

#include <algorithm>
#include <chrono>
#include <csignal>
#include <filesystem>
#include <fstream>
#include <nlohmann/json.hpp>
#include <sstream>
#include <string>
#include <thread>
#include <unistd.h>
#include <vector>

using json = nlohmann::json;
namespace fs = std::filesystem;

struct null {
    std::string ID;
    NLOHMANN_DEFINE_TYPE_INTRUSIVE(null, ID)
};

struct jobStatus {
    std::string jobName;
    int progress = 0;
    int phase = 0; // 0 for setup, 1 for execution, 2 for upload
    std::string status = "ACTIVE";
    NLOHMANN_DEFINE_TYPE_INTRUSIVE(jobStatus, jobName, progress, phase, status)
};

namespace Jobs {
    static void sendStatus(Api::Client& client, const jobStatus& status) {
        Api::Response res = client.Request("/jobs/status", "POST", status);

        if (!res) {
            Logger::Errln(std::format("status update failed (code {}): {}", res.statusCode, res.body));
        }
    }

    static int execProcess(const std::vector<std::string>& args, pid_t& outPid) {
        pid_t pid = fork();

        if (pid == 0) {
            std::vector<char*> argv;
            for (const auto& s : args)
                argv.push_back(const_cast<char*>(s.c_str()));
            argv.push_back(nullptr);

            execvp(argv[0], argv.data());
            _exit(127);
        }

        outPid = pid;
        return 0;
    }

    void Executor::Start() {
        m_requestThread = std::jthread(&Executor::requestLoop, this);
        m_execThread = std::jthread(&Executor::executionLoop, this);
    }

    void Executor::CancelActiveJob() {
        m_cancelJob.store(true);

        pid_t pid = m_activePid.load();
        if (pid > 0)
            kill(pid, SIGKILL);
    }

    void Executor::requestLoop() {
        using namespace std::chrono_literals;

        while (!m_cancelCtx->load()) {
            State state = m_state.load();

            switch (state) {
                case State::NoJobActive: {
                    Api::Response response = mr_client.Request("/jobs/next", "GET", null{});

                    if (response && response.statusCode != 204) {
                        json j = json::parse(response.body);

                        try {
                            m_jobSpec = j.get<JobSpec>();
                            m_progress.store(0);
                            m_jobStartTime = std::chrono::steady_clock::now();
                            m_state.store(State::JobActive);
                        } catch (const json::exception& e) {
                            Logger::Errln(std::format("parse error: {}", e.what()));
                        }
                    }
                    break;
                }

                case State::JobActive: {
                    sendStatus(mr_client, {.jobName = m_jobSpec.jobName, .progress = m_progress.load(), .phase = m_phase.load(), .status = "ACTIVE"});
                    break;
                }

                case State::JobSuccess: {
                    sendStatus(mr_client, {.jobName = m_jobSpec.jobName, .progress = m_progress.load(), .phase = m_phase.load(), .status = "SUCCESS"});
                    m_state.store(State::NoJobActive);
                    break;
                }

                case State::JobFail: {
                    sendStatus(mr_client, {.jobName = m_jobSpec.jobName, .progress = m_progress.load(), .phase = m_phase.load(), .status = "FAIL"});
                    m_state.store(State::NoJobActive);
                    break;
                }
            }

            std::this_thread::sleep_for(1s);
        }

        m_execThreadLatch.wait();
        mr_latch.count_down();
    }

    void Executor::executionLoop() {
        while (!m_cancelCtx->load()) {
            if (m_state.load() == State::JobActive) {
                JobSpec job;

                {
                    std::lock_guard lock(m_jobMutex);
                    job = m_jobSpec;
                }

                m_cancelJob.store(false);

                try {
                    runJob(job);

                    if (!m_cancelJob.load())
                        m_state.store(State::JobSuccess);
                    else
                        m_state.store(State::JobFail);

                } catch (const std::exception& e) {
                    Logger::Errln(std::format("execution error: {}", e.what()));
                    m_state.store(State::JobFail);
                }
            }

            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }

        m_execThreadLatch.count_down();
    }

    void Executor::downloadFile(const std::string& url, const std::string& outputPath) {
        Logger::Infoln(std::format("downloading from signed url -> {}", outputPath));

        pid_t pid;
        execProcess({"curl", "-L", "--fail", "-o", outputPath, url}, pid);

        int status;
        waitpid(pid, &status, 0);

        if (status != 0)
            throw std::runtime_error("download failed");
    }

    void Executor::uploadFile(const std::string& filePath, const std::string& signedUploadURL) {
        Logger::Infoln(std::format("uploading result -> {}", filePath));

        pid_t pid;
        execProcess({"curl", "-L", "--fail", "-X", "PUT", "--upload-file", filePath, signedUploadURL}, pid);

        int status;
        waitpid(pid, &status, 0);

        if (status != 0)
            throw std::runtime_error("upload failed");
    }

    void Executor::runJob(const JobSpec& job) {
        std::string jobID = std::to_string(std::time(nullptr));
        std::string dir = m_workspaceBase + "/" + jobID;

        fs::create_directories(dir + "/input");
        fs::create_directories(dir + "/output");

        m_progress.store(5);

        downloadFile(job.jar.url, dir + "/job.jar");

        int i = 0;
        for (const auto& input : job.data.input) {
            downloadFile(input, dir + "/input/input_" + std::to_string(i++));
        }

        m_progress.store(15);

        writeRunScript(job, dir);

        auto dockerArgs = buildDockerArgs(job, dir);

        m_phase.store(1);
        int result = runProcessStreaming(dockerArgs, job.timeoutSeconds);

        if (result == -1)
            throw std::runtime_error("timeout");

        if (result == -2)
            throw std::runtime_error("cancelled");

        m_progress.store(90);

        m_phase.store(2);

        runCommand({"zip", "-r", dir + "/output/result.zip", dir + "/output"});
        uploadFile(dir + "/output/result.zip", job.data.output);

        m_progress.store(100);

        fs::remove_all(dir);
    }

    int Executor::runProcessStreaming(const std::vector<std::string>& args, int timeoutSec) {
        int pipefd[2];
        pipe(pipefd);

        pid_t pid = fork();

        if (pid == 0) {
            dup2(pipefd[1], STDOUT_FILENO);
            dup2(pipefd[1], STDERR_FILENO);

            close(pipefd[0]);
            close(pipefd[1]);

            std::vector<char*> argv;
            for (const auto& s : args)
                argv.push_back(const_cast<char*>(s.c_str()));
            argv.push_back(nullptr);

            execvp(argv[0], argv.data());
            _exit(127);
        }

        close(pipefd[1]);
        m_activePid.store(pid);

        FILE* stream = fdopen(pipefd[0], "r");
        char buffer[1024];

        int elapsed = 0;
        int status = 0;

        while (true) {
            if (fgets(buffer, sizeof(buffer), stream)) {
                std::string line(buffer);

                parseProgress(line);
                Logger::Infoln(std::format("jar pritned \"{}\"", line));
            }

            pid_t result = waitpid(pid, &status, WNOHANG);
            if (result == pid)
                break;

            if (m_cancelJob.load()) {
                kill(pid, SIGKILL);
                waitpid(pid, &status, 0);
                fclose(stream);
                return -2;
            }

            if (elapsed++ >= timeoutSec * 10) {
                kill(pid, SIGKILL);
                waitpid(pid, &status, 0);
                fclose(stream);
                return -1;
            }

            std::this_thread::sleep_for(std::chrono::milliseconds(100));
        }

        fclose(stream);
        m_activePid.store(-1);

        return status;
    }

    void Executor::parseProgress(const std::string& line) {
        const std::string prefix = "PROGRESS:";

        if (line.rfind(prefix, 0) == 0) {
            try {
                int value = std::stoi(line.substr(prefix.size()));
                m_progress.store(std::clamp(value, 0, 100));
            } catch (...) {
            }
        }
    }

    void Executor::writeRunScript(const JobSpec& job, const std::string& dir) {
        std::ofstream f(dir + "/run.sh");

        f << "#!/bin/sh\n";
        f << "cd /workspace\n";

        std::string javaOpts = job.environment.count("JAVA_OPTS") ? job.environment.at("JAVA_OPTS") : "";

        f << "JAVA_OPTS=\"" << javaOpts << "\"\n";
        f << "stdbuf -oL java $JAVA_OPTS -cp job.jar " << job.jar.mainClass;

        for (const auto& arg : job.arguments)
            f << " " << arg;

        f << "\n";
        f.close();
        runCommand({"chmod", "+x", dir + "/run.sh"});
    }

    std::vector<std::string> Executor::buildDockerArgs(const JobSpec& job, const std::string& dir) {

        std::vector<std::string> args = {"docker",
                                         "run",
                                         "--rm",
                                         "--network=none",
                                         "--cpus=" + std::to_string(job.resources.executionCores),
                                         "--memory=" + std::to_string(job.resources.executionMemory) + "m",
                                         "--pids-limit=64",
                                         "--cap-drop=ALL",
                                         "--security-opt=no-new-privileges",
                                         "--read-only",
                                         "--tmpfs",
                                         "/tmp:rw,size=64m",
                                         "-v",
                                         dir + ":/workspace",
                                         "-w",
                                         "/workspace"};

        for (const auto& [k, v] : job.environment) {
            args.push_back("-e");
            args.push_back(k + "=" + v);
        }

        args.push_back("openjdk:17-jdk-slim");
        args.push_back("sh");
        args.push_back("/workspace/run.sh");

        return args;
    }

    void Executor::runCommand(const std::vector<std::string>& args) {
        pid_t pid;
        execProcess(args, pid);

        int status;
        waitpid(pid, &status, 0);

        if (status != 0)
            throw std::runtime_error("aws failed");
    }
} // namespace Jobs