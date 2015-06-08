'use strict';

var React = require('react'),
    Classable = require('./mixins/classable');

var TableRowItem = React.createClass({
  displayName: 'TableRowItem',

  mixins: [Classable],

  propTypes: {},

  getDefaultProps: function getDefaultProps() {
    return {};
  },

  render: function render() {
    var classes = this.getClasses('mui-table-rows-item');

    return React.createElement(
      'div',
      { className: classes },
      '(TableRowItem)',
      React.createElement(
        'div',
        { className: 'mui-table-rows-actions' },
        '(Actions)'
      )
    );
  }

});

module.exports = TableRowItem;