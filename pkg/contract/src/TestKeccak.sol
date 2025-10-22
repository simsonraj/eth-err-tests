// Original license: SPDX_License_Identifier: MIT
// SPDX-License-Identifier: MIT
pragma solidity ^0.8.0;

contract TestKeccak {
    function hash(uint32 iterations, bytes memory data) public pure returns (bytes32 result) {
        bytes32 h = keccak256(data);
        for (uint32 i = 0; i < iterations - 1; i++) {
            h = keccak256(abi.encode(h));
        }
        return h;
    }

    bytes32 currentHash;

    function hashAndStore(uint32 iterations, bytes memory data) public {
        currentHash = hash(iterations, data);
    }
}