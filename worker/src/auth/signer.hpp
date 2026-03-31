#pragma once
#include <string>

namespace Auth {
    std::string GenerateTimestamp();

    std::string SignRequest(const std::string& workerID,
                            const std::string& timestamp,
                            const std::string& secret);
} // namespace Auth
