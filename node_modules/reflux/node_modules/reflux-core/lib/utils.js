"use strict";

Object.defineProperty(exports, "__esModule", {
    value: true
});
exports.capitalize = capitalize;
exports.callbackName = callbackName;
exports.isObject = isObject;
exports.extend = extend;
exports.isFunction = isFunction;
exports.object = object;
exports.isArguments = isArguments;
exports.throwIf = throwIf;

function capitalize(string) {
    return string.charAt(0).toUpperCase() + string.slice(1);
}

function callbackName(string, prefix) {
    prefix = prefix || "on";
    return prefix + exports.capitalize(string);
}

var environment = {};

exports.environment = environment;
function checkEnv(target) {
    var flag = undefined;
    try {
        /*eslint-disable no-eval */
        if (eval(target)) {
            flag = true;
        }
        /*eslint-enable no-eval */
    } catch (e) {
        flag = false;
    }
    environment[callbackName(target, "has")] = flag;
}
checkEnv("setImmediate");
checkEnv("Promise");

/*
 * isObject, extend, isFunction, isArguments are taken from undescore/lodash in
 * order to remove the dependency
 */

function isObject(obj) {
    var type = typeof obj;
    return type === "function" || type === "object" && !!obj;
}

function extend(obj) {
    if (!isObject(obj)) {
        return obj;
    }
    var source, prop;
    for (var i = 1, length = arguments.length; i < length; i++) {
        source = arguments[i];
        for (prop in source) {
            if (Object.getOwnPropertyDescriptor && Object.defineProperty) {
                var propertyDescriptor = Object.getOwnPropertyDescriptor(source, prop);
                Object.defineProperty(obj, prop, propertyDescriptor);
            } else {
                obj[prop] = source[prop];
            }
        }
    }
    return obj;
}

function isFunction(value) {
    return typeof value === "function";
}

exports.EventEmitter = require("eventemitter3");

if (environment.hasSetImmediate) {
    exports.nextTick = function (callback) {
        setImmediate(callback);
    };
} else {
    exports.nextTick = function (callback) {
        setTimeout(callback, 0);
    };
}

function object(keys, vals) {
    var o = {},
        i = 0;
    for (; i < keys.length; i++) {
        o[keys[i]] = vals[i];
    }
    return o;
}

if (environment.hasPromise) {
    exports.Promise = Promise;
    exports.createPromise = function (resolver) {
        return new exports.Promise(resolver);
    };
} else {
    exports.Promise = null;
    exports.createPromise = function () {};
}

function isArguments(value) {
    return typeof value === "object" && "callee" in value && typeof value.length === "number";
}

function throwIf(val, msg) {
    if (val) {
        throw Error(msg || val);
    }
}