'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var ColorManipulator = require('./utils/color-manipulator');
var Typography = require('./styles/typography');
var EnhancedButton = require('./enhanced-button');

var FlatButton = React.createClass({
  displayName: 'FlatButton',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    className: React.PropTypes.string,
    disabled: React.PropTypes.bool,
    hoverColor: React.PropTypes.string,
    label: function label(props, propName, componentName) {
      if (!props.children && !props.label) {
        return new Error('Warning: Required prop `label` or `children` was not specified in `' + componentName + '`.');
      }
    },
    labelStyle: React.PropTypes.object,
    primary: React.PropTypes.bool,
    rippleColor: React.PropTypes.string,
    secondary: React.PropTypes.bool
  },

  getDefaultProps: function getDefaultProps() {
    return {
      labelStyle: {} };
  },

  getInitialState: function getInitialState() {
    return {
      hovered: false,
      isKeyboardFocused: false
    };
  },

  getThemeButton: function getThemeButton() {
    return this.context.muiTheme.component.button;
  },

  getTheme: function getTheme() {
    return this.context.muiTheme.component.flatButton;
  },

  _getColor: function _getColor() {
    var theme = this.getTheme();
    var color = this.props.disabled ? theme.disabledTextColor : this.props.primary ? theme.primaryTextColor : this.props.secondary ? theme.secondaryTextColor : theme.textColor;

    return {
      'default': color,
      hover: this.props.hoverColor || ColorManipulator.fade(ColorManipulator.lighten(color, 0.4), 0.15),
      ripple: this.props.rippleColor || ColorManipulator.fade(color, 0.8)
    };
  },

  getStyles: function getStyles() {
    var color = this._getColor();
    var styles = {
      root: {
        color: color['default'],
        transition: Transitions.easeOut(),
        fontSize: Typography.fontStyleButtonFontSize,
        letterSpacing: 0,
        textTransform: 'uppercase',
        fontWeight: Typography.fontWeightMedium,
        borderRadius: 2,
        userSelect: 'none',
        position: 'relative',
        overflow: 'hidden',
        backgroundColor: this.getTheme().color,
        lineHeight: this.getThemeButton().height + 'px',
        minWidth: this.getThemeButton().minWidth,
        padding: 0,
        margin: 0,
        //This is need so that ripples do not bleed past border radius.
        //See: http://stackoverflow.com/questions/17298739/css-overflow-hidden-not-working-in-chrome-when-parent-has-border-radius-and-chil
        transform: 'translate3d(0, 0, 0)' },
      label: {
        position: 'relative',
        padding: '0px ' + this.context.muiTheme.spacing.desktopGutterLess + 'px' },
      rootWhenHovered: {
        backgroundColor: color.hover
      },
      rippleColor: color.ripple
    };

    return styles;
  },

  render: function render() {
    var _props = this.props;
    var children = _props.children;
    var hoverColor = _props.hoverColor;
    var label = _props.label;
    var labelStyle = _props.labelStyle;
    var onBlur = _props.onBlur;
    var onMouseOut = _props.onMouseOut;
    var onMouseOver = _props.onMouseOver;
    var primary = _props.primary;
    var rippleColor = _props.rippleColor;
    var secondary = _props.secondary;
    var style = _props.style;

    var other = _objectWithoutProperties(_props, ['children', 'hoverColor', 'label', 'labelStyle', 'onBlur', 'onMouseOut', 'onMouseOver', 'primary', 'rippleColor', 'secondary', 'style']);

    var styles = this.getStyles();

    var labelElement;
    if (label) {
      labelElement = React.createElement(
        'span',
        { style: this.mergeAndPrefix(styles.label, this.props.labelStyle) },
        label
      );
    }

    return React.createElement(
      EnhancedButton,
      _extends({}, other, {
        ref: 'enhancedButton',
        style: this.mergeStyles(styles.root, (this.state.hovered || this.state.isKeyboardFocused) && !this.props.disabled && styles.rootWhenHovered, this.props.style),
        onMouseOver: this._handleMouseOver,
        onMouseOut: this._handleMouseOut,
        focusRippleColor: styles.rippleColor,
        touchRippleColor: styles.rippleColor,
        onKeyboardFocus: this._handleKeyboardFocus }),
      labelElement,
      this.props.children
    );
  },

  _handleMouseOver: function _handleMouseOver(e) {
    this.setState({ hovered: true });
    if (this.props.onMouseOver) {
      this.props.onMouseOver(e);
    }
  },

  _handleMouseOut: function _handleMouseOut(e) {
    this.setState({ hovered: false });
    if (this.props.onMouseOut) {
      this.props.onMouseOut(e);
    }
  },

  _handleKeyboardFocus: function _handleKeyboardFocus(e, isKeyboardFocused) {
    this.setState({ isKeyboardFocused: isKeyboardFocused });
  },

  _handleOnBlur: function _handleOnBlur(e) {
    this.setState({ hovered: false });
    if (this.props.onBlur) {
      this.props.onBlur(e);
    }
  }
});

module.exports = FlatButton;