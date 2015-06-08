'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react/addons');
var StylePropable = require('./mixins/style-propable');

var SvgIcon = React.createClass({
  displayName: 'SvgIcon',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  getTheme: function getTheme() {
    return this.context.muiTheme.palette;
  },

  getStyles: function getStyles() {
    return {
      display: 'inline-block',
      height: '24px',
      width: '24px',
      userSelect: 'none',
      fill: this.getTheme().textColor
    };
  },

  render: function render() {
    var _props = this.props;
    var viewBox = _props.viewBox;
    var style = _props.style;

    var other = _objectWithoutProperties(_props, ['viewBox', 'style']);

    return React.createElement(
      'svg',
      _extends({}, other, {
        viewBox: '0 0 24 24',
        style: this.mergeAndPrefix(this.getStyles(), this.props.style) }),
      this.props.children
    );
  }
});

module.exports = SvgIcon;