#include "logger.hpp"
#include <chrono>
#include <cstdlib>
#include <ctime>
#include <filesystem>
#include <fmt/format.h>
#include <iomanip>
#include <iostream>

#define CONSOLE_COLRESET "\033[0m"

#define CONSOLE_BLACK "\033[30m"
#define CONSOLE_RED "\033[31m"
#define CONSOLE_GREEN "\033[32m"
#define CONSOLE_YELLOW "\033[33m"
#define CONSOLE_BLUE "\033[34m"
#define CONSOLE_MAGENTA "\033[35m"
#define CONSOLE_CYAN "\033[36m"
#define CONSOLE_WHITE "\033[37m"

namespace Logging {
    auto Logger::Create(const std::string& fpath) -> std::expected<std::shared_ptr<Logger>, std::string> {
        if (std::filesystem::exists(fpath)) {
            const std::string base = std::filesystem::path(fpath).replace_extension().string();
            std::string command = fmt::format("gzip -c \"{}\" > \"{}_$(date +%Y-%m-%d_%H-%M-%S).log.gz\"", fpath, base);
            if (std::system(command.c_str()) == -1) {
                std::cout << "Failed to compress old log (is \"gzip\" installed?)\n";
            }
        }

        auto* ofstream = new std::ofstream(fpath, std::ios::out | std::ios::binary | std::ios::trunc);
        if (!ofstream->is_open()) {
            return std::unexpected("log file could not be opened");
        }

        return std::shared_ptr<Logger>(new Logger(ofstream));
    }

    Logger::Logger(std::ofstream* file) : m_ofstream(file) {}

    void Logger::Logln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S") << ' ' << s << '\n';
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S") << ' ' << s << '\n';
    }

    void Logger::Logln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S") << ' ' << s << '\n';
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S") << ' ' << s << '\n';
    }

    void Logger::Infoln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_CYAN << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [INFO] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [INFO] " << s << '\n';
    }

    void Logger::Infoln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_CYAN << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [INFO] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [INFO] " << s << '\n';
    }

    void Logger::Warnln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_YELLOW << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [WARN] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [WARN] " << s << '\n';
    }

    void Logger::Warnln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_YELLOW << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [WARN] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [WARN] " << s << '\n';
    }

    void Logger::Errln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_RED << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [ERROR] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [ERROR] " << s << '\n';
    }

    void Logger::Errln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_RED << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [ERROR] " << s << '\n' << CONSOLE_COLRESET;
        *m_ofstream << std::put_time(std::localtime(&now_t), "%Y/%m/%d %H:%M:%S") << " [ERROR] " << s << '\n';
    }
} // namespace Logging
