'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Spacing = require('./styles/spacing');
var Transitions = require('./styles/transitions');

var FontIcon = React.createClass({
  displayName: 'FontIcon',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    className: React.PropTypes.string,
    hoverColor: React.PropTypes.string
  },

  getInitialState: function getInitialState() {
    return {
      hovered: false };
  },

  getStyles: function getStyles() {
    var theme = this.context.muiTheme.palette;
    var styles = {
      position: 'relative',
      fontSize: Spacing.iconSize + 'px',
      display: 'inline-block',
      userSelect: 'none',
      transition: Transitions.easeOut()
    };

    if (!styles.color && !this.props.className) {
      styles.color = theme.textColor;
    }

    return styles;
  },

  render: function render() {
    var _props = this.props;
    var onMouseOut = _props.onMouseOut;
    var onMouseOver = _props.onMouseOver;
    var style = _props.style;

    var other = _objectWithoutProperties(_props, ['onMouseOut', 'onMouseOver', 'style']);

    var hoverStyle = this.props.hoverColor ? { color: this.props.hoverColor } : {};

    return React.createElement('span', _extends({}, other, {
      onMouseOut: this._handleMouseOut,
      onMouseOver: this._handleMouseOver,
      style: this.mergeAndPrefix(this.getStyles(), this.props.style, this.state.hovered && hoverStyle) }));
  },

  _handleMouseOut: function _handleMouseOut(e) {
    this.setState({ hovered: false });
    if (this.props.onMouseOut) {
      this.props.onMouseOut(e);
    }
  },

  _handleMouseOver: function _handleMouseOver(e) {
    this.setState({ hovered: true });
    if (this.props.onMouseOver) {
      this.props.onMouseOver(e);
    }
  }
});

module.exports = FontIcon;