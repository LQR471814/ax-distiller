// #include <iostream>
// // #include <fstream>
// #include <string>
// #include <cuco/static_map.cuh>
// #include <cuda_runtime.h>
// using namespace std;
#include <lib/hashmap.hpp>
#include <iostream>

#define CUDA_CHECK(call) \
    do { \
        cudaError_t error = call; \
        if (error != cudaSuccess) { \
            std::cerr << "CUDA error at " << __FILE__ << ":" << __LINE__ << " - " << cudaGetErrorString(error) << std::endl; \
            exit(1); \
        } \
    } while(0)
//custom value
int main() {
  //TO DO add a file system queue to proccess each file individually
  auto myFile = fopen("buffer.bin", "rb");
    if (myFile == nullptr) {
        cerr << "Failed to open file buffer.bin" << endl;
        return 1;
    }
    
    uint64_t headerBuf[2];
    auto n = fread(headerBuf, sizeof(uint64_t), 2, myFile);
    if (n != 2) {
        cerr << "Failed to read file header" << endl;
        fclose(myFile);
        return 1;
    }
    
    uint64_t length = headerBuf[0];
    uint64_t rootHash = headerBuf[1];
    
    cout << "File header:" << endl;
    cout << "  Length: " << length << " nodes" << endl;
    cout << "  Root hash: 0x" << hex << rootHash << dec << endl << endl;

    Node* buffer = new Node[length];  // Use dynamic allocation for large arrays
  return 0;
}