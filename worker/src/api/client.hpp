#pragma once
#include "../auth/signer.hpp"
#include "logger/logger.hpp"
#include <atomic>
#include <curl/curl.h>
#include <iostream>
#include <memory>
#include <mutex>
#include <nlohmann/json.hpp>
#include <optional>
#include <string_view>
#include <thread>

namespace Api {
    struct Response {
        long statusCode = 0;
        std::string body;
        CURLcode curlCode = CURLE_OK;

        explicit operator bool() const {
            return curlCode == CURLE_OK && statusCode >= 200 &&
                   statusCode < 300;
        }
    };

    class Client {
    public:
        Client(std::shared_ptr<std::atomic<bool>>& cancelCtx,
               std::string_view address, const std::string& workerID,
               std::string secretFPath);
        ~Client();

        template <typename T>
        Response Request(const std::string& route, const std::string& method,
                         const T& jsonData) {
            std::lock_guard<std::mutex> lock(m_requestMutex);

            Response resp;

            std::string url = std::string(m_address) + route;
            std::string ts = Auth::GenerateTimestamp();

            std::string body;
            if (method != "GET") {
                nlohmann::json j = jsonData;
                body = j.dump();
            }

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

            if (method != "GET") {
                headers = curl_slist_append(headers,
                                            "Content-Type: application/json");
            }

            headers = curl_slist_append(headers, "Expect:");

            curl_easy_reset(m_curl);

            std::string responseBody;

            curl_easy_setopt(m_curl, CURLOPT_URL, url.c_str());
            curl_easy_setopt(m_curl, CURLOPT_HTTPHEADER, headers);

            curl_easy_setopt(
                m_curl, CURLOPT_WRITEFUNCTION,
                +[](char* ptr, size_t size, size_t nmemb,
                    void* userdata) -> size_t {
                    auto* str = static_cast<std::string*>(userdata);
                    str->append(ptr, size * nmemb);
                    return size * nmemb;
                });

            curl_easy_setopt(m_curl, CURLOPT_WRITEDATA, &responseBody);

            if (method == "GET") {
                curl_easy_setopt(m_curl, CURLOPT_HTTPGET, 1L);
            } else {
                curl_easy_setopt(m_curl, CURLOPT_CUSTOMREQUEST, method.c_str());
                if (!body.empty()) {
                    curl_easy_setopt(m_curl, CURLOPT_COPYPOSTFIELDS,
                                     body.c_str());
                }
            }

            resp.curlCode = curl_easy_perform(m_curl);

            curl_easy_getinfo(m_curl, CURLINFO_RESPONSE_CODE, &resp.statusCode);

            curl_slist_free_all(headers);

            resp.body = std::move(responseBody);

            if (resp.curlCode != CURLE_OK) {
                Logger::Errln(std::format("request to {} failed: {}", route,
                                          curl_easy_strerror(resp.curlCode)));
            }

            return resp;
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
