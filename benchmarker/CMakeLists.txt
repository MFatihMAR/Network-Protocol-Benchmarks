cmake_minimum_required(VERSION 3.15)
project(benchmarker)

set(CMAKE_CXX_STANDARD 20)

add_executable(benchmarker main.cpp Proxy/Proxy.cpp)
