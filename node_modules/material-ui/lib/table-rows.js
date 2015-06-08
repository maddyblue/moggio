'use strict';

var React = require('react'),
    Classable = require('./mixins/classable'),
    TableRowsItem = require('./table-rows-item');

var TableRow = React.createClass({
  displayName: 'TableRow',

  mixins: [Classable],

  propTypes: {
    rowItems: React.PropTypes.array.isRequired
  },

  getDefaultProps: function getDefaultProps() {
    return {};
  },

  render: function render() {
    var classes = this.getClasses('mui-table-rows');

    return React.createElement(
      'div',
      { className: classes },
      this._getChildren()
    );
  },

  _getChildren: function _getChildren() {
    var children = [],
        rowItem,
        itemComponent;

    for (var i = 0; i < this.props.rowItems.length; i++) {
      rowItem = this.props.rowItems[i];

      /*
      for(var prop in rowItem) {
        if(rowItem.hasOwnProperty(prop)) {
          console.log(prop);
        }
      }
      console.log("--");
      */

      itemComponent = React.createElement(TableRowsItem, null);

      children.push(itemComponent);
    }

    return children;
  }

});

module.exports = TableRow;