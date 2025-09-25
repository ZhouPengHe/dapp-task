// SPDX-License-Identifier: MIT
pragma solidity ^0.8;

contract Counter {
    uint256 private count;
    event CountIncremented(uint256 newCount);
    constructor(uint256 init){
        count = init;
    }
    function increment() public {
        count += 1;
        emit CountIncremented(count);
    }
    function getCount() public view returns(uint256){
        return count;
    }
}