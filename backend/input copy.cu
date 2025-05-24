#include <iostream>
#include <fstream>
#include <string>
#include <cuco/static_map.cuh>
#include <cuda_runtime.h>
using namespace std;

//custom value
typedef struct Node {
    uint64_t fullkey;    //the key added up from all previous ancestors
    uint64_t hash;       //the current hash of this node and its subtree
    uint64_t firstChild;
    uint64_t nextSibling;
} Node;

// Helper function to check CUDA errors


int main() {
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
    
    // Allocate buffer for nodes
    Node* buffer = new Node[length];  // Use dynamic allocation for large arrays
    
    n = fread(buffer, sizeof(Node), length, myFile);
    if (n != length) {
        cerr << "Failed to read file contents. Expected " << length << " nodes, got " << n << endl;
        delete[] buffer;
        fclose(myFile);
        return 1;
    }
    
    fclose(myFile);
    cout << "Successfully read " << length << " nodes from file" << endl;
    
    // Copy to GPU
    Node* devicePtr;
    CUDA_CHECK(cudaMalloc(&devicePtr, length * sizeof(Node)));
    CUDA_CHECK(cudaMemcpy(devicePtr, buffer, length * sizeof(Node), cudaMemcpyHostToDevice));
    
    cout << "Data copied to GPU" << endl;
    
    // *** COPY BACK FROM GPU TO CPU ***
    Node* resultBuffer = new Node[length];
    CUDA_CHECK(cudaMemcpy(resultBuffer, devicePtr, length * sizeof(Node), cudaMemcpyDeviceToHost));
    
    cout << "Data copied back from GPU" << endl << endl;
    
    // *** PRINT FILE CONTENTS ***
    cout << "Node contents (showing first 10 nodes):" << endl;
    cout << "Index | FullKey            | Hash               | FirstChild | NextSibling" << endl;
    cout << "------|--------------------|--------------------|------------|------------" << endl;
    
    size_t maxPrint = min(static_cast<size_t>(10), static_cast<size_t>(length));
    for (size_t i = 0; i < maxPrint; ++i) {
        printf("%5zu | 0x%016lx | 0x%016lx | %10lu | %11lu\n", 
               i, 
               resultBuffer[i].fullkey, 
               resultBuffer[i].hash, 
               resultBuffer[i].firstChild, 
               resultBuffer[i].nextSibling);
    }
    
    if (length > 10) {
        cout << "... (showing only first 10 of " << length << " total nodes)" << endl;
    }
    
    // Print some statistics
    cout << endl << "Statistics:" << endl;
    uint64_t nonZeroFullkeys = 0;
    uint64_t nonZeroHashes = 0;
    uint64_t nodesWithChildren = 0;
    uint64_t nodesWithSiblings = 0;
    
    for (size_t i = 0; i < length; ++i) {
        if (resultBuffer[i].fullkey != 0) nonZeroFullkeys++;
        if (resultBuffer[i].hash != 0) nonZeroHashes++;
        if (resultBuffer[i].firstChild != 0) nodesWithChildren++;
        if (resultBuffer[i].nextSibling != 0) nodesWithSiblings++;
    }
    
    cout << "  Non-zero fullkeys: " << nonZeroFullkeys << " / " << length << endl;
    cout << "  Non-zero hashes: " << nonZeroHashes << " / " << length << endl;
    cout << "  Nodes with children: " << nodesWithChildren << " / " << length << endl;
    cout << "  Nodes with siblings: " << nodesWithSiblings << " / " << length << endl;
    
    // Verify data integrity (compare original vs copied-back data)
    bool dataMatches = true;
    for (size_t i = 0; i < length; ++i) {
        if (buffer[i].fullkey != resultBuffer[i].fullkey ||
            buffer[i].hash != resultBuffer[i].hash ||
            buffer[i].firstChild != resultBuffer[i].firstChild ||
            buffer[i].nextSibling != resultBuffer[i].nextSibling) {
            dataMatches = false;
            break;
        }
    }
    
    cout << endl << "Data integrity check: " << (dataMatches ? "PASSED" : "FAILED") << endl;
    
    // Cleanup
    delete[] buffer;
    delete[] resultBuffer;
    CUDA_CHECK(cudaFree(devicePtr));
    
    return 0;
}