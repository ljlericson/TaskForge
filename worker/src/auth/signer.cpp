#include <openssl/bio.h>
#include <openssl/evp.h>
#include <openssl/pem.h>
#include <openssl/rsa.h>

#include <chrono>
#include <iomanip>
#include <openssl/buffer.h>
#include <sstream>
#include <stdexcept>
#include <string>
#include <vector>

namespace Auth {

    static std::string Base64Encode(const unsigned char* buffer,
                                    size_t length) {
        BIO* bio = BIO_new(BIO_s_mem());
        BIO* b64 = BIO_new(BIO_f_base64());

        BIO_set_flags(b64, BIO_FLAGS_BASE64_NO_NL);
        bio = BIO_push(b64, bio);

        BIO_write(bio, buffer, (int)length);
        BIO_flush(bio);

        BUF_MEM* mem = nullptr;
        BIO_get_mem_ptr(bio, &mem);

        std::string result(mem->data, mem->length);

        BIO_free_all(bio);
        return result;
    }

    std::string GenerateTimestamp() {
        using namespace std::chrono;

        auto now = system_clock::now();
        auto seconds =
            duration_cast<std::chrono::seconds>(now.time_since_epoch());

        return std::to_string(seconds.count());
    }

    std::string Sha256Hex(const std::string& data) {
        EVP_MD_CTX* ctx = EVP_MD_CTX_new();
        if (!ctx)
            throw std::runtime_error("EVP_MD_CTX_new failed");

        unsigned char hash[EVP_MAX_MD_SIZE];
        unsigned int hashLen = 0;

        if (EVP_DigestInit_ex(ctx, EVP_sha256(), nullptr) != 1) {
            EVP_MD_CTX_free(ctx);
            throw std::runtime_error("EVP_DigestInit_ex failed");
        }

        if (EVP_DigestUpdate(ctx, data.data(), data.size()) != 1) {
            EVP_MD_CTX_free(ctx);
            throw std::runtime_error("EVP_DigestUpdate failed");
        }

        if (EVP_DigestFinal_ex(ctx, hash, &hashLen) != 1) {
            EVP_MD_CTX_free(ctx);
            throw std::runtime_error("EVP_DigestFinal_ex failed");
        }

        EVP_MD_CTX_free(ctx);

        std::ostringstream ss;
        ss << std::hex << std::setfill('0');

        for (unsigned int i = 0; i < hashLen; ++i) {
            ss << std::setw(2) << static_cast<int>(hash[i]);
        }

        return ss.str();
    }

    std::string SignRequest(const std::string& workerID,
                            const std::string& timestamp,
                            const std::string& method, const std::string& path,
                            const std::string& bodyHashHex,
                            const std::string& privateKeyPem) {

        std::string message = workerID + ":" + timestamp + ":" + method + ":" +
                              path + ":" + bodyHashHex;

        BIO* bio =
            BIO_new_mem_buf(privateKeyPem.data(), (int)privateKeyPem.size());

        if (!bio)
            throw std::runtime_error("BIO_new_mem_buf failed");

        EVP_PKEY* pkey =
            PEM_read_bio_PrivateKey(bio, nullptr, nullptr, nullptr);
        BIO_free(bio);

        if (!pkey)
            throw std::runtime_error("Failed to parse private key");

        EVP_MD_CTX* ctx = EVP_MD_CTX_new();
        if (!ctx) {
            EVP_PKEY_free(pkey);
            throw std::runtime_error("EVP_MD_CTX_new failed");
        }

        EVP_PKEY_CTX* pctx = nullptr;

        if (EVP_DigestSignInit(ctx, &pctx, EVP_sha256(), nullptr, pkey) <= 0) {
            EVP_MD_CTX_free(ctx);
            EVP_PKEY_free(pkey);
            throw std::runtime_error("EVP_DigestSignInit failed");
        }

        if (EVP_PKEY_CTX_set_rsa_padding(pctx, RSA_PKCS1_PADDING) <= 0) {
            EVP_MD_CTX_free(ctx);
            EVP_PKEY_free(pkey);
            throw std::runtime_error("Failed to set RSA padding");
        }

        if (EVP_DigestSignUpdate(ctx, message.data(), message.size()) <= 0) {
            EVP_MD_CTX_free(ctx);
            EVP_PKEY_free(pkey);
            throw std::runtime_error("EVP_DigestSignUpdate failed");
        }

        size_t sigLen = 0;

        if (EVP_DigestSignFinal(ctx, nullptr, &sigLen) <= 0) {
            EVP_MD_CTX_free(ctx);
            EVP_PKEY_free(pkey);
            throw std::runtime_error("EVP_DigestSignFinal(size) failed");
        }

        std::vector<unsigned char> signature(sigLen);

        if (EVP_DigestSignFinal(ctx, signature.data(), &sigLen) <= 0) {
            EVP_MD_CTX_free(ctx);
            EVP_PKEY_free(pkey);
            throw std::runtime_error("EVP_DigestSignFinal failed");
        }

        EVP_MD_CTX_free(ctx);
        EVP_PKEY_free(pkey);

        return Base64Encode(signature.data(), sigLen);
    }

} // namespace Auth