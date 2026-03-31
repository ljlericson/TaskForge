#pragma once
#include "../auth/signer.hpp"
#include "logger/logger.hpp"
#include <curl/curl.h>
#include <nlohmann/json.hpp>
#include <string_view>
#include <thread>

namespace Api {
    class Client {
    public:
        Client(std::string_view address, const std::string& workerID,
               std::string secretFPath);
        ~Client();

        template <typename T>
        void Request(const std::string& route, T jsonData) {
            nlohmann::json j = jsonData;
            std::string json = j.dump(4);
            std::string url = m_address.data() + route;
            std::string ts = Auth::GenerateTimestamp();

            Logger::Infoln(std::format("requesting addr {}", url).c_str());

            struct curl_slist* headers = NULL;

            headers = curl_slist_append(headers,
                                        ("X-Worker-ID: " + m_workerID).c_str());

            headers =
                curl_slist_append(headers, ("X-Timestamp: " + ts).c_str());

            headers = curl_slist_append(
                headers,
                ("X-Signature: " + Auth::SignRequest(m_workerID, ts, m_secret))
                    .c_str());

            curl_easy_setopt(m_curl, CURLOPT_URL, url.c_str());

            curl_easy_setopt(m_curl, CURLOPT_HTTPHEADER, headers);

            curl_easy_setopt(m_curl, CURLOPT_POSTFIELDS, json.c_str());

            CURLcode res = curl_easy_perform(m_curl);

            if (res != CURLE_OK) {
                Logger::Errln(
                    std::format("request failed: {}", curl_easy_strerror(res)));
            } else {
                Logger::Infoln("request successfull");
            }
            curl_slist_free_all(headers);
        }

        void RegisterWorker();

    private:
        void apiCall();

        std::string m_workerID;
        std::string m_secret;
        CURL* m_curl;
        std::jthread m_httpThread;
        std::string_view m_address;
    };
} // namespace Api
