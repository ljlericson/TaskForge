#include <fstream>
#include <iostream>
#include <nlohmann/json.hpp>

int main() {
  std::cout << "Hello World\n";
  nlohmann::json j;
  std::ifstream file("myjson.json");
  if (!file.is_open()) {
    return 1;
  }

  file >> j;
}
