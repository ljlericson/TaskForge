cmake_minimum_required(VERSION 3.30)

# Vcpkg toolchain
set(CMAKE_TOOLCHAIN_FILE "${CMAKE_CURRENT_SOURCE_DIR}/vcpkg/scripts/buildsystems/vcpkg.cmake"
    CACHE STRING "Vcpkg toolchain file")

project(TaskforgeWorker VERSION 1.0)

set(CMAKE_CXX_STANDARD 23)
set(CMAKE_CXX_STANDARD_REQUIRED True)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)

find_package(httplib CONFIG REQUIRED)
find_package(nlohmann_json CONFIG REQUIRED)

set(CMAKE_CXX_SCAN_FOR_MODULES OFF) 

add_executable(TaskforgeWorker)

file(GLOB_RECURSE TaskforgeWorker_SOURCES CONFIGURE_DEPENDS 
    "${CMAKE_CURRENT_SOURCE_DIR}/src/*.cpp"
    "${CMAKE_CURRENT_SOURCE_DIR}/src/*.h"
    "${CMAKE_CURRENT_SOURCE_DIR}/src/*.hpp"
    "${CMAKE_CURRENT_SOURCE_DIR}/src/*.inl"
)

target_sources(TaskforgeWorker PRIVATE ${TaskforgeWorker_SOURCES})

target_include_directories(TaskforgeWorker PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_link_libraries(TaskforgeWorker PRIVATE 
  httplib::httplib
  nlohmann_json::nlohmann_json
)

target_compile_options(TaskforgeWorker PRIVATE 
    -Wall 
    -Wextra 
    -Wpedantic
)
    
