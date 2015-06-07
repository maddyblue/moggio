'use strict';

var _createClass = (function () { function defineProperties(target, props) { for (var i = 0; i < props.length; i++) { var descriptor = props[i]; descriptor.enumerable = descriptor.enumerable || false; descriptor.configurable = true; if ('value' in descriptor) descriptor.writable = true; Object.defineProperty(target, descriptor.key, descriptor); } } return function (Constructor, protoProps, staticProps) { if (protoProps) defineProperties(Constructor.prototype, protoProps); if (staticProps) defineProperties(Constructor, staticProps); return Constructor; }; })();

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _classCallCheck(instance, Constructor) { if (!(instance instanceof Constructor)) { throw new TypeError('Cannot call a class as a function'); } }

/**
 * Copyright (c) 2015, Facebook, Inc.
 * All rights reserved.
 *
 * This source code is licensed under the BSD-style license found in the
 * LICENSE file in the root directory of this source tree. An additional grant
 * of patent rights can be found in the PATENTS file in the same directory.
 *
 * @providesModule ImmutableValue
 * @typechecks
 */

'use strict';

var invariant = require('./invariant');
var isNode = require('./isNode');
var keyOf = require('./keyOf');

var SECRET_KEY = keyOf({ _DONT_EVER_TYPE_THIS_SECRET_KEY: null });

/**
 * `ImmutableValue` provides a guarantee of immutability at developer time when
 * strict mode is used. The extra computations required to enforce immutability
 * are stripped out in production for performance reasons. `ImmutableValue`
 * guarantees to enforce immutability for enumerable, own properties. This
 * allows easy wrapping of `ImmutableValue` with the ability to store
 * non-enumerable properties on the instance that only your static methods
 * reason about. In order to achieve IE8 compatibility (which doesn't have the
 * ability to define non-enumerable properties), modules that want to build
 * their own reasoning of `ImmutableValue`s and store computations can define
 * their non-enumerable properties under the name `toString`, and in IE8 only
 * define a standard property called `toString` which will mistakenly be
 * considered not enumerable due to its name (but only in IE8). The only
 * limitation is that no one can store their own `toString` property.
 * https://developer.mozilla.org/en-US/docs/ECMAScript_DontEnum_attribute#JScript_DontEnum_Bug
 */

var ImmutableValue = (function () {
  /**
   * An instance of `ImmutableValue` appears to be a plain JavaScript object,
   * except `instanceof ImmutableValue` evaluates to `true`, and it is deeply
   * frozen in development mode.
   *
   * @param {number} secret Ensures this isn't accidentally constructed outside
   * of convenience constructors. If created outside of a convenience
   * constructor, may not be frozen. Forbidding that use case for now until we
   * have a better API.
   */

  function ImmutableValue(secret) {
    _classCallCheck(this, ImmutableValue);

    invariant(secret === ImmutableValue[SECRET_KEY], 'Only certain classes should create instances of `ImmutableValue`.' + 'You probably want something like ImmutableValueObject.create.');
  }

  _createClass(ImmutableValue, null, [{
    key: 'mergeAllPropertiesInto',

    /**
     * Helper method for classes that make use of `ImmutableValue`.
     * @param {ImmutableValue} destination Object to merge properties into.
     * @param {object} propertyObjects List of objects to merge into
     * `destination`.
     */
    value: function mergeAllPropertiesInto(destination, propertyObjects) {
      var argLength = propertyObjects.length;
      for (var i = 0; i < argLength; i++) {
        _extends(destination, propertyObjects[i]);
      }
    }
  }, {
    key: 'deepFreezeRootNode',

    /**
     * Freezes the supplied object deeply. Other classes may implement their own
     * version based on this.
     *
     * @param {*} object The object to freeze.
     */
    value: function deepFreezeRootNode(object) {
      if (isNode(object)) {
        return; // Don't try to freeze DOM nodes.
      }
      Object.freeze(object); // First freeze the object.
      for (var prop in object) {
        if (object.hasOwnProperty(prop)) {
          ImmutableValue.recurseDeepFreeze(object[prop]);
        }
      }
      Object.seal(object);
    }
  }, {
    key: 'recurseDeepFreeze',

    /**
     * Differs from `deepFreezeRootNode`, in that we first check if this is a
     * necessary recursion. If the object is already an `ImmutableValue`, then the
     * recursion is unnecessary as it is already frozen. That check obviously
     * wouldn't work for the root node version `deepFreezeRootNode`!
     */
    value: function recurseDeepFreeze(object) {
      if (isNode(object) || !ImmutableValue.shouldRecurseFreeze(object)) {
        return; // Don't try to freeze DOM nodes.
      }
      Object.freeze(object); // First freeze the object.
      for (var prop in object) {
        if (object.hasOwnProperty(prop)) {
          ImmutableValue.recurseDeepFreeze(object[prop]);
        }
      }
      Object.seal(object);
    }
  }, {
    key: 'shouldRecurseFreeze',

    /**
     * Checks if an object should be deep frozen. Instances of `ImmutableValue`
     * are assumed to have already been deep frozen, so we can have large
     * `process.env.NODE_ENV !== 'production'` time savings by skipping freezing of them.
     *
     * @param {*} object The object to check.
     * @return {boolean} Whether or not deep freeze is needed.
     */
    value: function shouldRecurseFreeze(object) {
      return typeof object === 'object' && !(object instanceof ImmutableValue) && object !== null;
    }
  }]);

  return ImmutableValue;
})();

ImmutableValue._DONT_EVER_TYPE_THIS_SECRET_KEY = Math.random();

module.exports = ImmutableValue;