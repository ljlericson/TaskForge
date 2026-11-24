#pragma once
#include <expected>
#include <fstream>
#include <memory>
#include <string>

namespace Logging {
    class Logger {
    public:
        static auto Create(const std::string& fpath) -> std::expected<std::shared_ptr<Logger>, std::string>;

        Logger(Logger&) = delete;
        Logger(Logger&&) = delete;

        void Logln(const char* s);
        void Logln(const std::string& s);

        void Infoln(const char* s);
        void Infoln(const std::string& s);

        void Warnln(const char* s);
        void Warnln(const std::string& s);

        void Errln(const char* s);
        void Errln(const std::string& s);

    private:
        Logger(std::ofstream* ofstream);

        std::unique_ptr<std::ofstream> m_ofstream = nullptr;
    };
} // namespace Logging
