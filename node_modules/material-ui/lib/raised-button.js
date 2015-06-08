'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var ColorManipulator = require('./utils/color-manipulator');
var Typography = require('./styles/typography');
var EnhancedButton = require('./enhanced-button');
var Paper = require('./paper');

var RaisedButton = React.createClass({
  displayName: 'RaisedButton',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    className: React.PropTypes.string,
    label: function label(props, propName, componentName) {
      if (!props.children && !props.label) {
        return new Error('Warning: Required prop `label` or `children` was not specified in `' + componentName + '`.');
      }
    },
    onMouseDown: React.PropTypes.func,
    onMouseUp: React.PropTypes.func,
    onMouseOut: React.PropTypes.func,
    onTouchEnd: React.PropTypes.func,
    onTouchStart: React.PropTypes.func,
    primary: React.PropTypes.bool,
    secondary: React.PropTypes.bool,
    labelStyle: React.PropTypes.object },

  getInitialState: function getInitialState() {
    var zDepth = this.props.disabled ? 0 : 1;
    return {
      zDepth: zDepth,
      initialZDepth: zDepth,
      hovered: false
    };
  },

  componentWillReceiveProps: function componentWillReceiveProps(nextProps) {
    var zDepth = nextProps.disabled ? 0 : 1;
    this.setState({
      zDepth: zDepth,
      initialZDepth: zDepth });
    this.styles = this.getStyles();
  },

  _getBackgroundColor: function _getBackgroundColor() {
    return this.props.disabled ? this.getTheme().disabledColor : this.props.primary ? this.getTheme().primaryColor : this.props.secondary ? this.getTheme().secondaryColor : this.getTheme().color;
  },

  _getLabelColor: function _getLabelColor() {
    return this.props.disabled ? this.getTheme().disabledTextColor : this.props.primary ? this.getTheme().primaryTextColor : this.props.secondary ? this.getTheme().secondaryTextColor : this.getTheme().textColor;
  },

  getThemeButton: function getThemeButton() {
    return this.context.muiTheme.component.button;
  },

  getTheme: function getTheme() {
    return this.context.muiTheme.component.raisedButton;
  },

  getStyles: function getStyles() {
    var amount = this.props.primary || this.props.secondary ? 0.4 : 0.08;
    var styles = {
      root: {
        display: 'inline-block',
        minWidth: this.getThemeButton().minWidth,
        height: this.getThemeButton().height,
        transition: Transitions.easeOut()
      },
      container: {
        position: 'relative',
        height: '100%',
        width: '100%',
        padding: 0,
        overflow: 'hidden',
        borderRadius: 2,
        transition: Transitions.easeOut(),
        backgroundColor: this._getBackgroundColor(),

        //This is need so that ripples do not bleed
        //past border radius.
        //See: http://stackoverflow.com/questions/17298739/css-overflow-hidden-not-working-in-chrome-when-parent-has-border-radius-and-chil
        transform: 'translate3d(0, 0, 0)'
      },
      label: {
        position: 'relative',
        opacity: 1,
        fontSize: '14px',
        letterSpacing: 0,
        textTransform: 'uppercase',
        fontWeight: Typography.fontWeightMedium,
        margin: 0,
        padding: '0px ' + this.context.muiTheme.spacing.desktopGutterLess + 'px',
        userSelect: 'none',
        lineHeight: this.getThemeButton().height + 'px',
        color: this._getLabelColor() },
      overlay: {
        transition: Transitions.easeOut(),
        top: 0
      },
      overlayWhenHovered: {
        backgroundColor: ColorManipulator.fade(this._getLabelColor(), amount)
      }
    };
    return styles;
  },

  render: function render() {
    var _props = this.props;
    var label = _props.label;
    var primary = _props.primary;
    var secondary = _props.secondary;

    var other = _objectWithoutProperties(_props, ['label', 'primary', 'secondary']);

    if (!this.hasOwnProperty('styles')) this.styles = this.getStyles();

    var labelElement;
    if (label) {
      labelElement = React.createElement(
        'span',
        { style: this.mergeAndPrefix(this.styles.label, this.props.labelStyle) },
        label
      );
    }

    var rippleColor = this.styles.label.color;
    var rippleOpacity = !(primary || secondary) ? 0.1 : 0.16;

    if (!this.hasOwnProperty('styles')) this.styles = this.getStyles();

    return React.createElement(
      Paper,
      {
        style: this.mergeAndPrefix(this.styles.root, this.props.style),
        zDepth: this.state.zDepth },
      React.createElement(
        EnhancedButton,
        _extends({}, other, {
          ref: 'container',
          style: this.mergeAndPrefix(this.styles.container),
          onMouseUp: this._handleMouseUp,
          onMouseDown: this._handleMouseDown,
          onMouseOut: this._handleMouseOut,
          onMouseOver: this._handleMouseOver,
          onTouchStart: this._handleTouchStart,
          onTouchEnd: this._handleTouchEnd,
          focusRippleColor: rippleColor,
          touchRippleColor: rippleColor,
          focusRippleOpacity: rippleOpacity,
          touchRippleOpacity: rippleOpacity,
          onKeyboardFocus: this._handleKeyboardFocus }),
        React.createElement(
          'div',
          { ref: 'overlay', style: this.mergeAndPrefix(this.styles.overlay, this.state.hovered && !this.props.disabled && this.styles.overlayWhenHovered) },
          labelElement,
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
      var amount = this.props.primary || this.props.secondary ? 0.4 : 0.08;
      React.findDOMNode(this.refs.overlay).style.backgroundColor = ColorManipulator.fade(this.mergeAndPrefix(this.styles.label, this.props.labelStyle).color, amount);
    } else if (!this.state.hovered) {
      this.setState({ zDepth: this.state.initialZDepth });
      React.findDOMNode(this.refs.overlay).style.backgroundColor = 'transparent';
    }
  } });

module.exports = RaisedButton;