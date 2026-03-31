#include <string>

namespace Logger {
    void Logln(const char* s);
    void Logln(const std::string& s);

    void Infoln(const char* s);
    void Logln(const std::string& s);

    void Warnln(const char* s);
    void Warnln(const std::string& s);

    void Errln(const char* s);
    void Errln(const std::string& s);
} // namespace Logger
