'use strict';

var React = require('react');

var TabTemplate = React.createClass({
  displayName: 'TabTemplate',

  render: function render() {

    var styles = {
      'display': 'block',
      'width': '100%',
      'position': 'relative',
      'text-align': 'initial'
    };

    return React.createElement(
      'div',
      { styles: styles },
      this.props.children
    );
  } });

module.exports = TabTemplate;