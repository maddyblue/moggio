/**
 * Copyright (c) 2015, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule PrefixIntervalTree
 * @typechecks
 */

'use strict';

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

/**
 * An interval tree that allows to set a number at index and given the value
 * find the largest index for which prefix sum is greater than or equal to value
 * (lower bound) or greater than value (upper bound)
 * Complexity:
 *   construct: O(n)
 *   query: O(log(n))
 *   memory: O(log(n)),
 * where n is leafCount from the constructor
 */

var PrefixIntervalTree = (function () {
  function PrefixIntervalTree( /*number*/leafCount, /*?number*/initialLeafValue) {
    _classCallCheck(this, PrefixIntervalTree);

    var internalLeafCount = this.getInternalLeafCount(leafCount);
    this._leafCount = leafCount;
    this._internalLeafCount = internalLeafCount;
    var nodeCount = 2 * internalLeafCount;
    var Int32Array = global.Int32Array || this._initArray;
    this._value = new Int32Array(nodeCount);
    this._initTables(initialLeafValue || 0);

    this.get = this.get.bind(this);
    this.set = this.set.bind(this);
    this.lowerBound = this.lowerBound.bind(this);
    this.upperBound = this.upperBound.bind(this);
  }

  _createClass(PrefixIntervalTree, [{
    key: 'getInternalLeafCount',
    value: function getInternalLeafCount( /*number*/leafCount) /*number*/{
      var internalLeafCount = 1;
      while (internalLeafCount < leafCount) {
        internalLeafCount *= 2;
      }
      return internalLeafCount;
    }
  }, {
    key: '_initArray',
    value: function _initArray( /*number*/size) /*array*/{
      var arr = [];
      while (size > 0) {
        size--;
        arr[size] = 0;
      }
      return arr;
    }
  }, {
    key: '_initTables',
    value: function _initTables( /*number*/initialLeafValue) {
      var firstLeaf = this._internalLeafCount;
      var lastLeaf = this._internalLeafCount + this._leafCount - 1;
      var i;
      for (i = firstLeaf; i <= lastLeaf; ++i) {
        this._value[i] = initialLeafValue;
      }
      var lastInternalNode = this._internalLeafCount - 1;
      for (i = lastInternalNode; i > 0; --i) {
        this._value[i] = this._value[2 * i] + this._value[2 * i + 1];
      }
    }
  }, {
    key: 'set',
    value: function set( /*number*/position, /*number*/value) {
      var nodeIndex = position + this._internalLeafCount;
      this._value[nodeIndex] = value;
      nodeIndex = Math.floor(nodeIndex / 2);
      while (nodeIndex !== 0) {
        this._value[nodeIndex] = this._value[2 * nodeIndex] + this._value[2 * nodeIndex + 1];
        nodeIndex = Math.floor(nodeIndex / 2);
      }
    }
  }, {
    key: 'get',

    /**
     * Returns an object {index, value} for given position (including value at
     * specified position), or the same for last position if provided position
     * is out of range
     */
    value: function get( /*number*/position) /*object*/{
      position = Math.min(position, this._leafCount);
      var nodeIndex = position + this._internalLeafCount;
      var result = this._value[nodeIndex];
      while (nodeIndex > 1) {
        if (nodeIndex % 2 === 1) {
          result = this._value[nodeIndex - 1] + result;
        }
        nodeIndex = Math.floor(nodeIndex / 2);
      }
      return { index: position, value: result };
    }
  }, {
    key: 'upperBound',

    /**
     * Returns an object {index, value} where index is index of leaf that was
     * found by upper bound algorithm. Upper bound finds first element for which
     * value is greater than argument
     */
    value: function upperBound( /*number*/value) /*object*/{
      var result = this._upperBoundImpl(1, 0, this._internalLeafCount - 1, value);
      if (result.index > this._leafCount - 1) {
        result.index = this._leafCount - 1;
      }
      return result;
    }
  }, {
    key: 'lowerBound',

    /**
     * Returns result in the same format as upperBound, but finds first element
     * for which value is greater than or equal to argument
     */
    value: function lowerBound( /*number*/value) /*object*/{
      var result = this.upperBound(value);
      if (result.value > value && result.index > 0) {
        var previousValue = result.value - this._value[this._internalLeafCount + result.index];
        if (previousValue === value) {
          result.value = previousValue;
          result.index--;
        }
      }
      return result;
    }
  }, {
    key: '_upperBoundImpl',
    value: function _upperBoundImpl(
    /*number*/nodeIndex,
    /*number*/nodeIntervalBegin,
    /*number*/nodeIntervalEnd,
    /*number*/value) /*object*/{
      if (nodeIntervalBegin === nodeIntervalEnd) {
        return {
          index: nodeIndex - this._internalLeafCount,
          value: this._value[nodeIndex] };
      }

      var nodeIntervalMidpoint = Math.floor((nodeIntervalBegin + nodeIntervalEnd + 1) / 2);
      if (value < this._value[nodeIndex * 2]) {
        return this._upperBoundImpl(2 * nodeIndex, nodeIntervalBegin, nodeIntervalMidpoint - 1, value);
      } else {
        var result = this._upperBoundImpl(2 * nodeIndex + 1, nodeIntervalMidpoint, nodeIntervalEnd, value - this._value[2 * nodeIndex]);
        result.value += this._value[2 * nodeIndex];
        return result;
      }
    }
  }]);

  return PrefixIntervalTree;
})();

module.exports = PrefixIntervalTree;