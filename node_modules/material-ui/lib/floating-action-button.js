'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var ColorManipulator = require('./utils/color-manipulator');
var EnhancedButton = require('./enhanced-button');
var FontIcon = require('./font-icon');
var Paper = require('./paper');

var getZDepth = function getZDepth(disabled) {
  var zDepth = disabled ? 0 : 2;
  return {
    zDepth: zDepth,
    initialZDepth: zDepth
  };
};

var RaisedButton = React.createClass({
  displayName: 'RaisedButton',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    iconClassName: React.PropTypes.string,
    iconStyle: React.PropTypes.object,
    mini: React.PropTypes.bool,
    onMouseDown: React.PropTypes.func,
    onMouseUp: React.PropTypes.func,
    onMouseOut: React.PropTypes.func,
    onTouchEnd: React.PropTypes.func,
    onTouchStart: React.PropTypes.func,
    secondary: React.PropTypes.bool
  },

  getInitialState: function getInitialState() {
    var zDepth = this.props.disabled ? 0 : 2;
    return {
      zDepth: zDepth,
      initialZDepth: zDepth,
      hovered: false };
  },

  componentWillMount: function componentWillMount() {
    this.setState(getZDepth(this.props.disabled));
  },

  componentWillReceiveProps: function componentWillReceiveProps(newProps) {
    if (newProps.disabled !== this.props.disabled) {
      this.setState(getZDepth(newProps.disabled));
    }
  },

  componentDidMount: function componentDidMount() {
    if (process.env.NODE_ENV !== 'production') {
      if (this.props.iconClassName && this.props.children) {
        var warning = 'You have set both an iconClassName and a child icon. ' + 'It is recommended you use only one method when adding ' + 'icons to FloatingActionButtons.';
        console.warn(warning);
      }
    }
  },

  _getBackgroundColor: function _getBackgroundColor() {
    return this.props.disabled ? this.getTheme().disabledColor : this.props.secondary ? this.getTheme().secondaryColor : this.getTheme().color;
  },

  getTheme: function getTheme() {
    return this.context.muiTheme.component.floatingActionButton;
  },

  _getIconColor: function _getIconColor() {
    return this.props.disabled ? this.getTheme().disabledTextColor : this.props.secondary ? this.getTheme().secondaryIconColor : this.getTheme().iconColor;
  },

  getStyles: function getStyles() {
    var styles = {
      root: {
        transition: Transitions.easeOut(),
        display: 'inline-block'
      },
      container: {
        transition: Transitions.easeOut(),
        position: 'relative',
        height: this.getTheme().buttonSize,
        width: this.getTheme().buttonSize,
        padding: 0,
        overflow: 'hidden',
        backgroundColor: this._getBackgroundColor(),
        borderRadius: '50%',
        textAlign: 'center',
        verticalAlign: 'bottom',
        //This is need so that ripples do not bleed
        //past border radius.
        //See: http://stackoverflow.com/questions/17298739/css-overflow-hidden-not-working-in-chrome-when-parent-has-border-radius-and-chil
        transform: 'translate3d(0, 0, 0)'
      },
      icon: {
        lineHeight: this.getTheme().buttonSize + 'px',
        fill: this.getTheme().iconColor,
        color: this._getIconColor()
      },
      overlay: {
        transition: Transitions.easeOut(),
        top: 0
      },
      containerWhenMini: {
        height: this.getTheme().miniSize,
        width: this.getTheme().miniSize
      },
      iconWhenMini: {
        lineHeight: this.getTheme().miniSize + 'px'
      },
      overlayWhenHovered: {
        backgroundColor: ColorManipulator.fade(this._getIconColor(), 0.4)
      }
    };
    return styles;
  },

  render: function render() {
    var _props = this.props;
    var icon = _props.icon;
    var mini = _props.mini;
    var secondary = _props.secondary;

    var other = _objectWithoutProperties(_props, ['icon', 'mini', 'secondary']);

    var styles = this.getStyles();

    var iconElement;
    if (this.props.iconClassName) {
      iconElement = React.createElement(FontIcon, {
        className: this.props.iconClassName,
        style: this.mergeAndPrefix(styles.icon, this.props.mini && styles.iconWhenMini, this.props.iconStyle) });
    }

    var rippleColor = styles.icon.color;

    return React.createElement(
      Paper,
      {
        style: this.mergeAndPrefix(styles.root, this.props.style),
        zDepth: this.state.zDepth,
        circle: true },
      React.createElement(
        EnhancedButton,
        _extends({}, other, {
          ref: 'container',
          style: this.mergeAndPrefix(styles.container, this.props.mini && styles.containerWhenMini),
          onMouseDown: this._handleMouseDown,
          onMouseUp: this._handleMouseUp,
          onMouseOut: this._handleMouseOut,
          onMouseOver: this._handleMouseOver,
          onTouchStart: this._handleTouchStart,
          onTouchEnd: this._handleTouchEnd,
          focusRippleColor: rippleColor,
          touchRippleColor: rippleColor,
          onKeyboardFocus: this._handleKeyboardFocus }),
        React.createElement(
          'div',
          {
            ref: 'overlay',
            style: this.mergeAndPrefix(styles.overlay, this.state.hovered && !this.props.disabled && styles.overlayWhenHovered) },
          iconElement,
          this.props.children
        )
      )
    );
  },

  _handleMouseDown: function _handleMouseDown(e) {
    //only listen to left clicks
    if (e.button === 0) {
      this.setState({ zDepth: this.state.initialZDepth + 1 });
    }
    if (this.props.onMouseDown) this.props.onMouseDown(e);
  },

  _handleMouseUp: function _handleMouseUp(e) {
    this.setState({ zDepth: this.state.initialZDepth });
    if (this.props.onMouseUp) this.props.onMouseUp(e);
  },

  _handleMouseOut: function _handleMouseOut(e) {
    if (!this.refs.container.isKeyboardFocused()) this.setState({ zDepth: this.state.initialZDepth, hovered: false });
    if (this.props.onMouseOut) this.props.onMouseOut(e);
  },

  _handleMouseOver: function _handleMouseOver(e) {
    if (!this.refs.container.isKeyboardFocused()) this.setState({ hovered: true });
    if (this.props.onMouseOver) this.props.onMouseOver(e);
  },

  _handleTouchStart: function _handleTouchStart(e) {
    this.setState({ zDepth: this.state.initialZDepth + 1 });
    if (this.props.onTouchStart) this.props.onTouchStart(e);
  },

  _handleTouchEnd: function _handleTouchEnd(e) {
    this.setState({ zDepth: this.state.initialZDepth });
    if (this.props.onTouchEnd) this.props.onTouchEnd(e);
  },

  _handleKeyboardFocus: function _handleKeyboardFocus(e, keyboardFocused) {
    if (keyboardFocused && !this.props.disabled) {
      this.setState({ zDepth: this.state.initialZDepth + 1 });
      React.findDOMNode(this.refs.overlay).style.backgroundColor = ColorManipulator.fade(this.getStyles().icon.color, 0.4);
    } else if (!this.state.hovered) {
      this.setState({ zDepth: this.state.initialZDepth });
      React.findDOMNode(this.refs.overlay).style.backgroundColor = 'transparent';
    }
  } });

module.exports = RaisedButton;