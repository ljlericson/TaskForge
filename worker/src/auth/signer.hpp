#pragma once
#include <string>

namespace Auth {
    std::string GenerateTimestamp();

    std::string Base64Encode(const unsigned char* buffer, size_t length);

    std::string Sha256Hex(const std::string& data);

    std::string SignRequest(const std::string& workerID,
                            const std::string& timestamp,
                            const std::string& method, const std::string& path,
                            const std::string& bodyHashHex,
                            const std::string& privateKeyPem);
} // namespace Auth
