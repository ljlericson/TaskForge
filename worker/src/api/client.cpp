#include "client.hpp"
#include <curl/easy.h>
#include <fmt/format.h>
#include <fstream>
#include <logger/logger.hpp>

namespace Api {
    Client::Client(std::shared_ptr<std::atomic<bool>>& cancelCtx, std::string_view address, const std::string& workerID, std::string secretFPath)
        : m_cancelCtx(cancelCtx), m_workerID(workerID), m_address(address) {
        curl_global_init(CURL_GLOBAL_ALL);
        m_curl = curl_easy_init();

        std::ifstream file(secretFPath);
        if (!file.is_open()) {
            // to do: make Logger::Fatalln()
            Logger::Errln("private key file could not be opened, filepath: " + secretFPath);
            exit(-1);
        }

        std::string line;
        while (std::getline(file, line)) {
            m_secret.append(line);
            m_secret.push_back('\n');
        }
    }

    Client::~Client() {
        curl_global_cleanup();
        curl_easy_cleanup(m_curl);
    }

    struct registerRequest {
        std::string id;
        NLOHMANN_DEFINE_TYPE_INTRUSIVE(registerRequest, id)
    };

    void Client::RegisterWorker() {
        auto res = this->Request("/workers/register", "POST", registerRequest{.id = m_workerID});

        if (!res) {
            Logger::Errln(fmt::format("registration failed, server gave error: {}", res.body).c_str());
            m_cancelCtx->store(true);
        }
    }
} // namespace Api
