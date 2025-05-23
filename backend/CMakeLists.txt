cmake_minimum_required(VERSION 3.18)  # CUDA support requires 3.18+
project(distiller LANGUAGES CXX CUDA)

include(./CPM.cmake)

# Find CUDA
find_package(CUDA REQUIRED)

# Set CUDA standard
set(CMAKE_CUDA_STANDARD 17)
set(CMAKE_CUDA_STANDARD_REQUIRED ON)

CPMAddPackage(
  NAME cuco
  GITHUB_REPOSITORY NVIDIA/cuCollections
  GIT_TAG dev
  OPTIONS
     "BUILD_TESTS OFF"
     "BUILD_BENCHMARKS OFF"
     "BUILD_EXAMPLES OFF"
)

# Add executable with CUDA files
add_executable(distiller
    input.cu
)

# # Link CUDA runtime
# target_link_libraries(distiller2 ${CUDA_LIBRARIES})