#include "client.hpp"
#include <curl/easy.h>
#include <fmt/format.h>
#include <fstream>

namespace Api {
    auto Client::Create(std::shared_ptr<std::atomic<bool>>& cancelCtx, const std::shared_ptr<Logging::Logger>& logger, std::string_view address, const std::string& workerID, std::string secretFPath)
        -> std::expected<std::shared_ptr<Client>, std::string> {
        std::ifstream file(secretFPath);
        if (!file.is_open()) {
            return std::unexpected("private key file could not be opened, filepath: " + secretFPath);
        }

        return std::shared_ptr<Client>(new Client(file, cancelCtx, logger, address, workerID));
    }

    Client::Client(std::ifstream& file, std::shared_ptr<std::atomic<bool>>& cancelCtx, const std::shared_ptr<Logging::Logger>& logger, std::string_view address, const std::string& workerID)
        : m_cancelCtx(cancelCtx), m_logger(logger), m_workerID(workerID), m_address(address) {
        curl_global_init(CURL_GLOBAL_ALL);
        m_curl = curl_easy_init();

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

    Response Client::RegisterWorker() { return this->Request("/workers/register", "POST", registerRequest{.id = m_workerID}); }
} // namespace Api
