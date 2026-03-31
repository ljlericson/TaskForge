#include "client.hpp"
#include <curl/easy.h>
#include <fstream>
#include <logger/logger.hpp>

// kinda stupic i know, but also isn't it stupid how c++ has no
// built in panic?
#define PANIC() exit(-1)

namespace Api {
    Client::Client(std::string_view address, const std::string& workerID,
                   std::string secretFPath)
        : m_workerID(workerID), m_address(address) {
        curl_global_init(CURL_GLOBAL_ALL);
        m_curl = curl_easy_init();

        std::ifstream file(secretFPath);
        if (!file.is_open()) {
            Logger::Errln("private key file could not be opened");
            PANIC();
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
        this->Request("/workers/register", registerRequest{.id = m_workerID});
    }
} // namespace Api
