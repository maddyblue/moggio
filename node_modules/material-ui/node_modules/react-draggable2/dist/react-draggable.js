(function webpackUniversalModuleDefinition(root, factory) {
	if(typeof exports === 'object' && typeof module === 'object')
		module.exports = factory(require("React"));
	else if(typeof define === 'function' && define.amd)
		define(["React"], factory);
	else if(typeof exports === 'object')
		exports["ReactDraggable"] = factory(require("React"));
	else
		root["ReactDraggable"] = factory(root["React"]);
})(this, function(__WEBPACK_EXTERNAL_MODULE_1__) {
return /******/ (function(modules) { // webpackBootstrap
/******/ 	// The module cache
/******/ 	var installedModules = {};
/******/
/******/ 	// The require function
/******/ 	function __webpack_require__(moduleId) {
/******/
/******/ 		// Check if module is in cache
/******/ 		if(installedModules[moduleId])
/******/ 			return installedModules[moduleId].exports;
/******/
/******/ 		// Create a new module (and put it into the cache)
/******/ 		var module = installedModules[moduleId] = {
/******/ 			exports: {},
/******/ 			id: moduleId,
/******/ 			loaded: false
/******/ 		};
/******/
/******/ 		// Execute the module function
/******/ 		modules[moduleId].call(module.exports, module, module.exports, __webpack_require__);
/******/
/******/ 		// Flag the module as loaded
/******/ 		module.loaded = true;
/******/
/******/ 		// Return the exports of the module
/******/ 		return module.exports;
/******/ 	}
/******/
/******/
/******/ 	// expose the modules object (__webpack_modules__)
/******/ 	__webpack_require__.m = modules;
/******/
/******/ 	// expose the module cache
/******/ 	__webpack_require__.c = installedModules;
/******/
/******/ 	// __webpack_public_path__
/******/ 	__webpack_require__.p = "";
/******/
/******/ 	// Load entry module and return exports
/******/ 	return __webpack_require__(0);
/******/ })
/************************************************************************/
/******/ ([
/* 0 */
/***/ function(module, exports, __webpack_require__) {

	'use strict';
	
	var React = __webpack_require__(1);
	var emptyFunction = function () {};
	
	// for accessing browser globals
	var root = typeof window !== 'undefined' ? window : this;
	var bodyElement;
	if (typeof document !== 'undefined' && 'body' in document) {
	  bodyElement = document.body;
	}
	
	function updateBoundState (state, bound) {
	  if (!bound) return state;
	  bound = String(bound);
	  var boundTop = !!~bound.indexOf('top');
	  var boundRight = !!~bound.indexOf('right');
	  var boundBottom = !!~bound.indexOf('bottom');
	  var boundLeft = !!~bound.indexOf('left');
	  var boundAll = !!~bound.indexOf('all') ||
	    !(boundTop || boundRight || boundBottom || boundLeft);
	  var boundBox = !~bound.indexOf('point');
	  state.boundTop = boundAll || boundTop;
	  state.boundRight = boundAll || boundRight;
	  state.boundBottom = boundAll || boundBottom;
	  state.boundLeft = boundAll || boundLeft;
	  state.boundBox = boundBox;
	  return state;
	};
	
	function createUIEvent(draggable) {
	  return {
	    position: {
	      top: draggable.state.offsetTop,
	      left: draggable.state.offsetLeft
	    }
	  };
	}
	
	function canDragY(draggable) {
	  return draggable.props.axis === 'both' ||
	      draggable.props.axis === 'y';
	}
	
	function canDragX(draggable) {
	  return draggable.props.axis === 'both' ||
	      draggable.props.axis === 'x';
	}
	
	function isFunction(func) {
	  return typeof func === 'function' || Object.prototype.toString.call(func) === '[object Function]'
	}
	
	// @credits https://gist.github.com/rogozhnikoff/a43cfed27c41e4e68cdc
	function findInArray(array, callback) {
	  for (var i = 0, length = array.length, element = null; i < length, element = array[i]; i++) {
	    if (callback.apply(callback, [element, i, array])) return element;
	  }
	}
	
	function matchesSelector(el, selector) {
	  var method = findInArray([
	    'matches',
	    'webkitMatchesSelector',
	    'mozMatchesSelector',
	    'msMatchesSelector',
	    'oMatchesSelector'
	  ], function(method){
	    return isFunction(el[method]);
	  });
	
	  return el[method].call(el, selector);
	}
	
	// @credits: http://stackoverflow.com/questions/4817029/whats-the-best-way-to-detect-a-touch-screen-device-using-javascript/4819886#4819886
	var isTouchDevice = 'ontouchstart' in root // works on most browsers
	                 || 'onmsgesturechange' in root; // works on ie10 on ms surface
	
	// look ::handleDragStart
	//function isMultiTouch(e) {
	//  return e.touches && Array.isArray(e.touches) && e.touches.length > 1
	//}
	
	/**
	 * simple abstraction for dragging events names
	 * */
	var dragEventFor = (function () {
	  var eventsFor = {
	    touch: {
	      start: 'touchstart',
	      move: 'touchmove',
	      end: 'touchend'
	    },
	    mouse: {
	      start: 'mousedown',
	      move: 'mousemove',
	      end: 'mouseup'
	    }
	  };
	  return eventsFor[isTouchDevice ? 'touch' : 'mouse'];
	})();
	
	/**
	 * get {clientX, clientY} positions of control
	 * */
	function getControlPosition(e) {
	  var position = (e.touches && e.touches[0]) || e;
	  return {
	    clientX: position.clientX,
	    clientY: position.clientY
	  }
	}
	
	function addEvent(el, event, handler) {
	  if (!el) { return; }
	  if (el.attachEvent) {
	    el.attachEvent('on' + event, handler);
	  } else if (el.addEventListener) {
	    el.addEventListener(event, handler, true);
	  } else {
	    el['on' + event] = handler;
	  }
	}
	
	function removeEvent(el, event, handler) {
	  if (!el) { return; }
	  if (el.detachEvent) {
	    el.detachEvent('on' + event, handler);
	  } else if (el.removeEventListener) {
	    el.removeEventListener(event, handler, true);
	  } else {
	    el['on' + event] = null;
	  }
	}
	
	module.exports = React.createClass({
	  displayName: 'Draggable',
	  mixins: [React.addons.PureRenderMixin],
	
	  propTypes: {
	    /**
	     * `axis` determines which axis the draggable can move.
	     *
	     * 'both' allows movement horizontally and vertically.
	     * 'x' limits movement to horizontal axis.
	     * 'y' limits movement to vertical axis.
	     *
	     * Defaults to 'both'.
	     */
	    axis: React.PropTypes.oneOf(['both', 'x', 'y']),
	
	    /**
	     * `handle` specifies a selector to be used as the handle that initiates drag.
	     *
	     * Example:
	     *
	     * ```jsx
	     *  var App = React.createClass({
	     *      render: function () {
	     *        return (
	     *          <Draggable handle=".handle">
	     *            <div>
	     *                <div className="handle">Click me to drag</div>
	     *                <div>This is some other content</div>
	     *            </div>
	     *          </Draggable>
	     *        );
	     *      }
	     *  });
	     * ```
	     */
	    handle: React.PropTypes.string,
	
	    /**
	     * `cancel` specifies a selector to be used to prevent drag initialization.
	     *
	     * Example:
	     *
	     * ```jsx
	     *  var App = React.createClass({
	     *      render: function () {
	     *          return(
	     *              <Draggable cancel=".cancel">
	     *                  <div>
	     *                    <div className="cancel">You can't drag from here</div>
	     *            <div>Dragging here works fine</div>
	     *                  </div>
	     *              </Draggable>
	     *          );
	     *      }
	     *  });
	     * ```
	     */
	    cancel: React.PropTypes.string,
	
	    /**
	     * `bound` determines whether to bound the movement to the parent box.
	     *
	     * The property takes a list of space-separated strings. The Draggable
	     * is bounded by the nearest DOMNode.offsetParent. To set the offset
	     * parent, give it a position value other than 'static'.
	     *
	     * Optionally choose one or more bounds from:
	     * 'top' bounds movement to the top edge of the parent box.
	     * 'right' bounds movement to the right edge of the parent box.
	     * 'bottom' bounds movement to the bottom edge of the parent box.
	     * 'left' bounds movement to the left edge of the parent box.
	     * 'all' bounds movement to all edges (default if not specified).
	     *
	     * Optionally choose one anchor from:
	     * 'point' to constrain only the top-left corner.
	     * 'box' to constrain the entire box (default if not specified).
	     *
	     * You may use more than one bound, e.g. 'top left point'. Set to a
	     * falsy value to disable.
	     *
	     * Defaults to 'all box'.
	     */
	    bound: React.PropTypes.string,
	
	    /**
	     * `grid` specifies the x and y that dragging should snap to.
	     *
	     * Example:
	     *
	     * ```jsx
	     *   var App = React.createClass({
	     *       render: function () {
	     *           return (
	     *              <Draggable grid={[25, 25]}>
	     *                   <div>I snap to a 25 x 25 grid</div>
	     *               </Draggable>
	     *           );
	     *      }
	     *   });
	     * ```
	     */
	    grid: React.PropTypes.arrayOf(React.PropTypes.number),
	
	    /**
	     * `constrain` takes a function to constrain the dragging.
	     *
	     * Example:
	     *
	     * ```jsx
	     *   function constrain (snap) {
	     *         function constrainOffset (offset, prev) {
	     *               var delta = offset - prev;
	     *               if (Math.abs(delta) >= snap) {
	     *                     return prev + (delta < 0 ? -snap : snap);
	     *               }
	     *               return prev;
	     *         }
	     *         return function (pos) {
	     *               return {
	     *                     top: constrainOffset(pos.top, pos.prevTop),
	     *                     left: constrainOffset(pos.left, pos.prevLeft)
	     *               };
	     *         };
	     *   }
	     *   var App = React.createClass({
	     *       render: function () {
	     *           return (
	     *               <Draggable constrain={constrain}>
	     *                   <div>I snap to a 25 x 25 grid</div>
	     *               </Draggable>
	     *           );
	     *       }
	     *   });
	     * ```
	     */
	    constrain: React.PropTypes.func,
	
	    /**
	     * `start` specifies the x and y that the dragged item should start at
	     *
	     * Example:
	     *
	     * ```jsx
	     *  var App = React.createClass({
	     *      render: function () {
	     *          return (
	     *              <Draggable start={{x: 25, y: 25}}>
	     *                  <div>I start with left: 25px; top: 25px;</div>
	     *              </Draggable>
	     *          );
	     *      }
	     *  });
	     * ```
	     */
	    start: React.PropTypes.object,
	
	    /**
	     * `zIndex` specifies the zIndex to use while dragging.
	     *
	     * Example:
	     *
	     * ```jsx
	     *  var App = React.createClass({
	     *      render: function () {
	     *          return (
	     *              <Draggable zIndex={100}>
	     *                  <div>I have a zIndex</div>
	     *              </Draggable>
	     *          );
	     *      }
	     *  });
	     * ```
	     */
	    zIndex: React.PropTypes.number,
	
	    /**
	     * `useChild` determines whether to use the first child as root.
	     *
	     * If false, a div is created. This option is required if any children
	     * have a ref.
	     *
	     * Defaults to true.
	     */
	    useChild: React.PropTypes.bool,
	
	    /**
	     * Called when dragging starts.
	     *
	     * Example:
	     *
	     * ```js
	     *  function (event, ui) {}
	     * ```
	     *
	     * `event` is the Event that was triggered.
	     * `ui` is an object:
	     *
	     * ```js
	     *  {
	     *    position: {top: 0, left: 0}
	     *  }
	     * ```
	     */
	    onStart: React.PropTypes.func,
	
	    /**
	     * Called while dragging.
	     *
	     * Example:
	     *
	     * ```js
	     *  function (event, ui) {}
	     * ```
	     *
	     * `event` is the Event that was triggered.
	     * `ui` is an object:
	     *
	     * ```js
	     *  {
	     *    position: {top: 0, left: 0}
	     *  }
	     * ```
	     */
	    onDrag: React.PropTypes.func,
	
	    /**
	     * Called when dragging stops.
	     *
	     * Example:
	     *
	     * ```js
	     *  function (event, ui) {}
	     * ```
	     *
	     * `event` is the Event that was triggered.
	     * `ui` is an object:
	     *
	     * ```js
	     *  {
	     *    position: {top: 0, left: 0}
	     *  }
	     * ```
	     */
	    onStop: React.PropTypes.func,
	
	    /**
	     * A workaround option which can be passed if onMouseDown needs to be accessed, since it'll always be blocked (due to that there's internal use of onMouseDown)
	     *
	     */
	    onMouseDown: React.PropTypes.func
	  },
	
	  getDefaultProps: function () {
	    return {
	      axis: 'both',
	      bound: null,
	      handle: null,
	      cancel: null,
	      grid: null,
	      start: {},
	      zIndex: NaN,
	      useChild: true,
	      onStart: emptyFunction,
	      onDrag: emptyFunction,
	      onStop: emptyFunction,
	      onMouseDown: emptyFunction
	    };
	  },
	
	  getInitialState: function () {
	    var state = {
	      // Whether or not currently dragging
	      dragging: false,
	
	      // Pointer offset on screen
	      clientX: 0, clientY: 0,
	
	      // DOMNode offset relative to parent
	      offsetLeft: this.props.start.x || 0, offsetTop: this.props.start.y || 0
	    };
	
	    updateBoundState(state, this.props.bound);
	
	    return state;
	  },
	
	  componentWillReceiveProps: function (nextProps) {
	    var state = updateBoundState({}, nextProps.bound);
	    if (nextProps.start) {
	      if (nextProps.start.x != null) {
	        state.offsetLeft = nextProps.start.x || 0;
	      }
	      if (nextProps.start.y != null) {
	        state.offsetTop = nextProps.start.y || 0;
	      }
	    }
	    this.setState(state);
	  },
	
	  componentWillUnmount: function() {
	    // Remove any leftover event handlers
	    removeEvent(root, dragEventFor['move'], this.handleDrag);
	    removeEvent(root, dragEventFor['end'], this.handleDragEnd);
	  },
	
	  handleDragStart: function (e) {
	    // todo: write right implementation to prevent multitouch drag
	    // prevent multi-touch events
	    // if (isMultiTouch(e)) {
	    //     this.handleDragEnd.apply(e, arguments);
	    //     return
	    // }
	
	    // Make it possible to attach event handlers on top of this one
	    this.props.onMouseDown(e);
	
	    // Short circuit if handle or cancel prop was provided and selector doesn't match
	    if ((this.props.handle && !matchesSelector(e.target, this.props.handle)) ||
	      (this.props.cancel && matchesSelector(e.target, this.props.cancel))) {
	      return;
	    }
	
	    var dragPoint = getControlPosition(e);
	
	    // Initiate dragging
	    this.setState({
	      dragging: true,
	      clientX: dragPoint.clientX,
	      clientY: dragPoint.clientY
	    });
	
	    // Call event handler
	    this.props.onStart(e, createUIEvent(this));
	
	    // Add event handlers
	    addEvent(root, dragEventFor['move'], this.handleDrag);
	    addEvent(root, dragEventFor['end'], this.handleDragEnd);
	
	    // Add dragging class to body element
	    if (bodyElement) bodyElement.className += ' react-draggable-dragging';
	  },
	
	  handleDragEnd: function (e) {
	    // Short circuit if not currently dragging
	    if (!this.state.dragging) {
	      return;
	    }
	
	    // Turn off dragging
	    this.setState({
	      dragging: false
	    });
	
	    // Call event handler
	    this.props.onStop(e, createUIEvent(this));
	
	    // Remove event handlers
	    removeEvent(root, dragEventFor['move'], this.handleDrag);
	    removeEvent(root, dragEventFor['end'], this.handleDragEnd);
	
	    // Remove dragging class from body element
	    if (bodyElement) {
	      var className = bodyElement.className;
	      bodyElement.className =
	        className.replace(/(?:^|\s+)react-draggable-dragging\b/, ' ');
	    }
	  },
	
	  handleDrag: function (e) {
	    var dragPoint = getControlPosition(e);
	    var offsetLeft = this._toPixels(this.state.offsetLeft);
	    var offsetTop = this._toPixels(this.state.offsetTop);
	
	    var state = {
	      offsetLeft: offsetLeft,
	      offsetTop: offsetTop
	    };
	
	    // Get parent DOM node
	    var node = this.getDOMNode();
	    var offsetParent = node.offsetParent;
	    var offset, boundingValue;
	
	    if (canDragX(this)) {
	      // Calculate updated position
	      offset = offsetLeft + dragPoint.clientX - this.state.clientX;
	
	      // Bound movement to parent box
	      if (this.state.boundLeft) {
	        boundingValue = state.offsetLeft - node.offsetLeft;
	        if (offset < boundingValue) {
	          offset = boundingValue;
	        }
	      }
	      if (this.state.boundRight) {
	        boundingValue += offsetParent.clientWidth;
	        if (this.state.boundBox) {
	          boundingValue -= node.offsetWidth;
	        }
	        if (offset > boundingValue) {
	          offset = boundingValue;
	        }
	      }
	      // Update left
	      state.offsetLeft = offset;
	    }
	
	    if (canDragY(this)) {
	      // Calculate updated position
	      offset = offsetTop + dragPoint.clientY - this.state.clientY;
	      // Bound movement to parent box
	      if (this.state.boundTop) {
	        boundingValue = state.offsetTop - node.offsetTop;
	        if (offset < boundingValue) {
	          offset = boundingValue;
	        }
	      }
	      if (this.state.boundBottom) {
	        boundingValue += offsetParent.clientHeight;
	        if (this.state.boundBox) {
	          boundingValue -= node.offsetHeight;
	        }
	        if (offset > boundingValue) {
	          offset = boundingValue;
	        }
	      }
	      // Update top
	      state.offsetTop = offset;
	    }
	
	    var constrain = this.props.constrain;
	    var grid = this.props.grid;
	
	    // Backwards-compatibility for snap to grid
	    if (!constrain && Array.isArray(grid)) {
	      var constrainOffset = function (offset, prev, snap) {
	        var delta = offset - prev;
	        if (Math.abs(delta) >= snap) {
	          return prev + parseInt(delta / snap, 10) * snap;
	        }
	        return prev;
	      };
	      constrain = function (pos) {
	        return {
	          left: constrainOffset(pos.left, pos.prevLeft, grid[0]),
	          top: constrainOffset(pos.top, pos.prevTop, grid[1])
	        };
	      };
	    }
	
	    // Constrain if function has been provided
	    var positions;
	    if (constrain) {
	      // Constrain positions
	      positions = constrain({
	        prevLeft: this.state.offsetLeft,
	        prevTop: this.state.offsetTop,
	        left: state.offsetLeft,
	        top: state.offsetTop
	      });
	      if (positions) {
	        // Update left
	        if ('left' in positions && !isNaN(positions.left)) {
	          state.offsetLeft = positions.left;
	        }
	        // Update top
	        if ('top' in positions && !isNaN(positions.top)) {
	          state.offsetTop = positions.top;
	        }
	      }
	    }
	
	    // Save new state
	    state.clientX = this.state.clientX + (state.offsetLeft - offsetLeft);
	    state.clientY = this.state.clientY + (state.offsetTop - offsetTop);
	    this.setState(state);
	
	    // Call event handler
	    this.props.onDrag(e, createUIEvent(this));
	  },
	
	  onTouchStart: function (e) {
	    e.preventDefault(); // prevent for scroll
	    return this.handleDragStart.apply(this, arguments);
	  },
	
	  render: function () {
	    var style = {
	      top: this.state.offsetTop,
	      left: this.state.offsetLeft
	    };
	
	    // Set zIndex if currently dragging and prop has been provided
	    if (this.state.dragging && !isNaN(this.props.zIndex)) {
	      style.zIndex = this.props.zIndex;
	    }
	
	    var props = {
	      style: style,
	      className: 'react-draggable',
	
	      onMouseDown: this.handleDragStart,
	      onTouchStart: this.onTouchStart,
	
	      onMouseUp: this.handleDragEnd,
	      onTouchEnd: this.handleDragEnd
	    };
	
	    // Reuse the child provided
	    // This makes it flexible to use whatever element is wanted (div, ul, etc)
	    if (this.props.useChild) {
	      return React.addons.cloneWithProps(React.Children.only(this.props.children), props);
	    }
	
	    return React.DOM.div(props, this.props.children);
	  },
	
	  _toPixels: function (value) {
	
	    // Support percentages
	    if (typeof value == 'string' && value.slice(-1) == '%') {
	      return parseInt((+value.replace('%', '') / 100) *
	        this.getDOMNode().offsetParent.clientWidth, 10) || 0;
	    }
	
	    // Invalid values become zero
	    var i = parseInt(value, 10);
	    if (isNaN(i) || !isFinite(i)) return 0;
	
	    return i;
	  }
	
	});


/***/ },
/* 1 */
/***/ function(module, exports, __webpack_require__) {

	module.exports = __WEBPACK_EXTERNAL_MODULE_1__;

/***/ }
/******/ ])
});

//# sourceMappingURL=react-draggable.js.map