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

  propTypes: {
    viewBox: React.PropTypes.string
  },

  getDefaultProps: function getDefaultProps() {
    return {
      viewBox: '0 0 24 24'
    };
  },

  render: function render() {
    var _props = this.props;
    var viewBox = _props.viewBox;
    var style = _props.style;

    var other = _objectWithoutProperties(_props, ['viewBox', 'style']);

    var mergedStyles = this.mergeAndPrefix({
      display: 'inline-block',
      height: '24px',
      width: '24px',
      userSelect: 'none',
      fill: this.context.muiTheme.palette.textColor
    }, style);

    return React.createElement(
      'svg',
      _extends({}, other, {
        viewBox: viewBox,
        style: mergedStyles }),
      this.props.children
    );
  }
});

module.exports = SvgIcon;