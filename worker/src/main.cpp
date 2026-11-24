#include "app/app.hpp"
#include <iostream>

int main() {
    auto app = App::Worker::Create("config/worker.json");
    if (!app) {
        std::cout << app.error() << std::endl;
    } else {
        app.value()->Run();
    }
}
