'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var TextField = require('./text-field');
var DropDownMenu = require('./drop-down-menu');

var SelectField = React.createClass({
  displayName: 'SelectField',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    errorText: React.PropTypes.string,
    floatingLabelText: React.PropTypes.string,
    hintText: React.PropTypes.string,
    id: React.PropTypes.string,
    multiLine: React.PropTypes.bool,
    onBlur: React.PropTypes.func,
    onChange: React.PropTypes.func,
    onFocus: React.PropTypes.func,
    onKeyDown: React.PropTypes.func,
    onEnterKeyDown: React.PropTypes.func,
    type: React.PropTypes.string,
    rows: React.PropTypes.number,
    inputStyle: React.PropTypes.object,
    floatingLabelStyle: React.PropTypes.object,
    autoWidth: React.PropTypes.bool,
    menuItems: React.PropTypes.array.isRequired,
    menuItemStyle: React.PropTypes.object,
    selectedIndex: React.PropTypes.number
  },

  getDefaultProps: function getDefaultProps() {
    return {};
  },

  getStyles: function getStyles() {
    var styles = {
      selectfield: {
        root: {
          height: 'auto',
          position: 'relative',
          width: '100%'
        },
        label: {
          paddingLeft: 0,
          top: 4,
          width: '100%'
        },
        icon: {
          top: 20,
          right: 0
        },
        underline: {
          borderTop: 'none'
        }
      }
    };
    return styles;
  },

  onChange: function onChange(e, index, payload) {
    e.target.value = payload;
    if (this.props.onChange) this.props.onChange(e);
  },

  render: function render() {
    var styles = this.getStyles();
    return React.createElement(
      TextField,
      this.props,
      React.createElement(DropDownMenu, _extends({}, this.props, {
        onChange: this.onChange,
        style: styles.selectfield.root,
        labelStyle: styles.selectfield.label,
        iconStyle: styles.selectfield.icon,
        underlineStyle: styles.selectfield.underline,
        autoWidth: false
      }))
    );
  }
});

module.exports = SelectField;