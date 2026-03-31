#include "logger.hpp"
#include <chrono>
#include <ctime>
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

namespace Logger {
    void Logln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << '['
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n';
    }

    void Logln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << '['
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n';
    }

    void Infoln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_CYAN << "[INFO "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }

    void Infoln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_CYAN << "[INFO "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }

    void Warnln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_YELLOW << "[WARN "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }

    void Warnln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_YELLOW << "[WARN "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }

    void Errln(const char* s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_RED << "[ERR  "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }

    void Errln(const std::string& s) {
        auto now = std::chrono::system_clock::now();

        std::time_t now_t = std::chrono::system_clock::to_time_t(now);
        std::cout << CONSOLE_RED << "[ERR  "
                  << std::put_time(std::localtime(&now_t), "%Y-%m-%d %H:%M:%S")
                  << "] " << s << '\n'
                  << CONSOLE_COLRESET;
    }
} // namespace Logger
