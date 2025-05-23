#include <iostream>
#include <fstream> 
#include <string>
using namespace std; 

typedef struct Node {
  uint64_t key;
  uint64_t firstChild;
  uint64_t nextSibling;
};



int main() {
  auto myFile = fopen("buffer.bin", "rb");
  if (myFile == nullptr) {
    return 1;
  }

  uint64_t headerBuf[2];
  auto n = fread(headerBuf, sizeof(uint64_t), 2, myFile);
  if (n != 2) {
    throw "failed to read file header";
  }

  uint64_t length = headerBuf[0];
  uint64_t rootHash = headerBuf[1];
  uint64_t buffer[length * 24];

  auto n = fread(buffer, sizeof(Node), length, myFile);
  if (n != length) {
    throw "failed to read file contents";
  }

  return 0;
}

