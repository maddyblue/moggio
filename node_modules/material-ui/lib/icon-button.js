'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var EnhancedButton = require('./enhanced-button');
var FontIcon = require('./font-icon');
var Tooltip = require('./tooltip');

var IconButton = React.createClass({
  displayName: 'IconButton',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    className: React.PropTypes.string,
    disabled: React.PropTypes.bool,
    iconClassName: React.PropTypes.string,
    iconStyle: React.PropTypes.object,
    onBlur: React.PropTypes.func,
    onFocus: React.PropTypes.func,
    tooltip: React.PropTypes.string,
    touch: React.PropTypes.bool
  },

  getInitialState: function getInitialState() {
    return {
      tooltipShown: false
    };
  },

  getDefaultProps: function getDefaultProps() {
    return {
      iconStyle: {}
    };
  },

  componentDidMount: function componentDidMount() {
    if (this.props.tooltip) {
      this._positionTooltip();
    }
    if (process.env.NODE_ENV !== 'production') {
      if (this.props.iconClassName && this.props.children) {
        var warning = 'You have set both an iconClassName and a child icon. ' + 'It is recommended you use only one method when adding ' + 'icons to IconButtons.';
        console.warn(warning);
      }
    }
  },

  getStyles: function getStyles() {
    var spacing = this.context.muiTheme.spacing;
    var palette = this.context.muiTheme.palette;

    var styles = {
      root: {
        position: 'relative',
        boxSizing: 'border-box',
        transition: Transitions.easeOut(),
        padding: spacing.iconSize / 2,
        width: spacing.iconSize * 2,
        height: spacing.iconSize * 2
      },
      tooltip: {
        boxSizing: 'border-box',
        marginTop: this.context.muiTheme.component.button.iconButtonSize + 4
      },
      icon: {
        color: palette.textColor,
        fill: palette.textColor
      },
      overlay: {
        position: 'relative',
        top: 0,
        width: '100%',
        height: '100%',
        background: palette.disabledColor
      },
      disabled: {
        color: palette.disabledColor,
        fill: palette.disabledColor
      }
    };
    return styles;
  },

  render: function render() {
    var _props = this.props;
    var disabled = _props.disabled;
    var iconClassName = _props.iconClassName;
    var tooltip = _props.tooltip;
    var touch = _props.touch;

    var other = _objectWithoutProperties(_props, ['disabled', 'iconClassName', 'tooltip', 'touch']);

    var fonticon;

    var styles = this.getStyles();

    var tooltipElement = tooltip ? React.createElement(Tooltip, {
      ref: 'tooltip',
      label: tooltip,
      show: this.state.tooltipShown,
      touch: touch,
      style: this.mergeStyles(styles.tooltip) }) : null;

    if (iconClassName) {
      var _props$iconStyle = this.props.iconStyle;
      var iconHoverColor = _props$iconStyle.iconHoverColor;

      var iconStyle = _objectWithoutProperties(_props$iconStyle, ['iconHoverColor']);

      fonticon = React.createElement(FontIcon, {
        className: iconClassName,
        hoverColor: disabled ? null : iconHoverColor,
        style: this.mergeStyles(styles.icon, disabled ? styles.disabled : {}, iconStyle) });
    }

    var children = disabled ? this._addStylesToChildren(styles.disabled) : this.props.children;

    return React.createElement(
      EnhancedButton,
      _extends({}, other, {
        ref: 'button',
        centerRipple: true,
        disabled: disabled,
        style: this.mergeStyles(styles.root, this.props.style),
        onBlur: this._handleBlur,
        onFocus: this._handleFocus,
        onMouseOut: this._handleMouseOut,
        onMouseOver: this._handleMouseOver,
        onKeyboardFocus: this._handleKeyboardFocus }),
      tooltipElement,
      fonticon,
      children
    );
  },

  _addStylesToChildren: function _addStylesToChildren(styles) {
    var children = [];

    React.Children.forEach(this.props.children, function (child) {
      children.push(React.cloneElement(child, {
        key: child.props.key ? child.props.key : children.length,
        style: styles
      }));
    });

    return children;
  },

  _positionTooltip: function _positionTooltip() {
    var tooltip = React.findDOMNode(this.refs.tooltip);
    var tooltipWidth = tooltip.offsetWidth;
    var buttonWidth = 48;

    tooltip.style.left = (tooltipWidth - buttonWidth) / 2 * -1 + 'px';
  },

  _showTooltip: function _showTooltip() {
    if (!this.props.disabled && this.props.tooltip) {
      this.setState({ tooltipShown: true });
    }
  },

  _hideTooltip: function _hideTooltip() {
    if (this.props.tooltip) this.setState({ tooltipShown: false });
  },

  _handleBlur: function _handleBlur(e) {
    this._hideTooltip();
    if (this.props.onBlur) this.props.onBlur(e);
  },

  _handleFocus: function _handleFocus(e) {
    this._showTooltip();
    if (this.props.onFocus) this.props.onFocus(e);
  },

  _handleMouseOut: function _handleMouseOut(e) {
    if (!this.refs.button.isKeyboardFocused()) this._hideTooltip();
    if (this.props.onMouseOut) this.props.onMouseOut(e);
  },

  _handleMouseOver: function _handleMouseOver(e) {
    this._showTooltip();
    if (this.props.onMouseOver) this.props.onMouseOver(e);
  },

  _handleKeyboardFocus: function _handleKeyboardFocus(e, keyboardFocused) {
    if (keyboardFocused && !this.props.disabled) {
      this._showTooltip();
      if (this.props.onFocus) this.props.onFocus(e);
    } else if (!this.state.hovered) {
      this._hideTooltip();
      if (this.props.onBlur) this.props.onBlur(e);
    }
  }

});

module.exports = IconButton;