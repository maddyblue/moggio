'use strict';

var React = require('react');
var SvgIcon = require('../svg-icon');

var DropDownArrow = React.createClass({
  displayName: 'DropDownArrow',

  render: function render() {
    return React.createElement(
      SvgIcon,
      this.props,
      React.createElement('polygon', { points: '7,9.5 12,14.5 17,9.5 ' })
    );
  }

});

module.exports = DropDownArrow;