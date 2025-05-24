#include <hashmap.hpp>
#include <iostream>

HashMap init(uint64_t length, Node* buffer){
    //assumes length is the true length of buffer
    if(buffer == nullptr){
        throw("invalid input buffer please try again");
        return {};
    }

    Node* devicePtr;
    cudaMalloc(&devicePtr, length * sizeof(Node));
    cudaMemcpy(devicePtr, buffer, length, cudaMemcpyHostToDevice);
    HashMap ret = {length, devicePtr};
    return ret;
}

uint64_t hash(HashMap map, uint64_t key) { //we hash based on the fullkey 
        return key % map.length;
}

Node lookUp(HashMap map, uint64_t fullKey){
    return map.arr[fullKey];
}