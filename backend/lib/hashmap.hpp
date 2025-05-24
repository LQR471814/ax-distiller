#ifndef HASHMAP_H
#define HASHMAP_H
#include <vector>
#include <cstdint>
#include <cuda_runtime.h>

typedef struct Node {
    uint64_t fullkey;    //the key added up from all previous ancestors
    uint64_t hash;       //the current hash of this node and its subtree
    uint64_t firstChild;
    uint64_t nextSibling;
} Node;

typedef struct HashMap {
    uint64_t length;
    Node* arr;
} HashMap;

HashMap init(uint64_t length, Node* buffer);
Node lookUp(uint64_t fullKey);

#endif HASHMAP_H