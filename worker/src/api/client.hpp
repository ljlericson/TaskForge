#pragma once
#include "../auth/signer.hpp"
#include "logger/logger.hpp"
#include <atomic>
#include <curl/curl.h>
#include <memory>
#include <mutex>
#include <nlohmann/json.hpp>
#include <string_view>
#include <thread>

namespace Api {
    class Client {
    public:
        Client(std::shared_ptr<std::atomic<bool>>& cancelCtx,
               std::string_view address, const std::string& workerID,
               std::string secretFPath);
        ~Client();

        template <typename T>
        bool Request(const std::string& route, const std::string& method,
                     const T& jsonData) {
            std::lock_guard<std::mutex> lock(m_requestMutex);

            nlohmann::json j = jsonData;
            std::string body = j.dump();

            std::string url = std::string(m_address) + route;
            std::string ts = Auth::GenerateTimestamp();

            std::string bodyHashHex = Auth::Sha256Hex(body);

            std::string signature = Auth::SignRequest(
                m_workerID, ts, method, route, bodyHashHex, m_secret);

            struct curl_slist* headers = nullptr;

            headers = curl_slist_append(headers,
                                        ("X-Worker-ID: " + m_workerID).c_str());
            headers =
                curl_slist_append(headers, ("X-Timestamp: " + ts).c_str());
            headers = curl_slist_append(headers,
                                        ("X-Signature: " + signature).c_str());
            headers =
                curl_slist_append(headers, "Content-Type: application/json");

            // prevents 100-continue issues
            headers = curl_slist_append(headers, "Expect:");

            curl_easy_reset(m_curl);

            curl_easy_setopt(m_curl, CURLOPT_URL, url.c_str());
            curl_easy_setopt(m_curl, CURLOPT_HTTPHEADER, headers);

            curl_easy_setopt(m_curl, CURLOPT_CUSTOMREQUEST, method.c_str());

            if (!body.empty()) {
                curl_easy_setopt(m_curl, CURLOPT_COPYPOSTFIELDS, body.c_str());
            }

            Logger::Warnln("BODY: " + body);
            Logger::Warnln("SIZE: " + std::to_string(body.size()));

            CURLcode res = curl_easy_perform(m_curl);

            curl_slist_free_all(headers);

            if (res != CURLE_OK) {
                Logger::Errln(std::format("request to {} failed: {}", route,
                                          curl_easy_strerror(res)));

                return false;
            }

            return true;
        }

        void RegisterWorker();

    private:
        void apiCall();

        std::shared_ptr<std::atomic<bool>> m_cancelCtx;
        std::string m_workerID;
        std::string m_secret;
        CURL* m_curl;
        std::mutex m_requestMutex;
        std::jthread m_httpThread;
        std::string_view m_address;
    };
} // namespace Api
