#include "signer.hpp"
#include <chrono>
#include <iomanip>
#include <openssl/hmac.h>
#include <sstream>

namespace Auth {
    std::string toHex(const unsigned char* data, size_t len) {
        std::stringstream ss;

        for (size_t i = 0; i < len; i++)
            ss << std::hex << std::setw(2) << std::setfill('0') << (int)data[i];

        return ss.str();
    }

    std::string GenerateTimestamp() {
        using namespace std::chrono;

        auto now = system_clock::now();
        auto seconds =
            duration_cast<std::chrono::seconds>(now.time_since_epoch());

        return std::to_string(seconds.count());
    }

    std::string SignRequest(const std::string& workerID,
                            const std::string& timestamp,
                            const std::string& secret) {
        std::string message = workerID + ":" + timestamp;

        unsigned char result[EVP_MAX_MD_SIZE];
        unsigned int len = 0;

        HMAC(EVP_sha256(), secret.data(), secret.size(),
             (unsigned char*)message.data(), message.size(), result, &len);

        return toHex(result, len);
    }
} // namespace Auth
