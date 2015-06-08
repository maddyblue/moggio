'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var Spacing = require('./styles/spacing');
var ClickAwayable = require('./mixins/click-awayable');
var KeyLine = require('./utils/key-line');
var Paper = require('./paper');
var FontIcon = require('./font-icon');
var Menu = require('./menu/menu');

var DropDownIcon = React.createClass({
  displayName: 'DropDownIcon',

  mixins: [StylePropable, ClickAwayable],

  propTypes: {
    onChange: React.PropTypes.func,
    menuItems: React.PropTypes.array.isRequired,
    closeOnMenuItemClick: React.PropTypes.bool,
    iconStyle: React.PropTypes.object,
    iconClassName: React.PropTypes.string },

  getInitialState: function getInitialState() {
    return {
      open: false };
  },

  getDefaultProps: function getDefaultProps() {
    return {
      closeOnMenuItemClick: true
    };
  },

  componentClickAway: function componentClickAway() {
    this.setState({ open: false });
  },

  getStyles: function getStyles() {
    var iconWidth = 48;
    var styles = {
      root: {
        display: 'inline-block',
        width: iconWidth + 'px !important',
        position: 'relative',
        height: Spacing.desktopToolbarHeight,
        fontSize: Spacing.desktopDropDownMenuFontSize,
        cursor: 'pointer'
      },
      menu: {
        transition: Transitions.easeOut(),
        right: '-14px !important',
        top: '9px !important',
        opacity: this.props.open ? 1 : 0
      },
      menuItem: { // similair to drop down menu's menu item styles
        paddingRight: Spacing.iconSize + Spacing.desktopGutterLess * 2,
        height: Spacing.desktopDropDownMenuItemHeight,
        lineHeight: Spacing.desktopDropDownMenuItemHeight + 'px'
      }
    };
    return styles;
  },

  render: function render() {
    var _props = this.props;
    var style = _props.style;
    var children = _props.children;
    var menuItems = _props.menuItems;
    var closeOnMenuItemClick = _props.closeOnMenuItemClick;
    var iconStyle = _props.iconStyle;
    var iconClassName = _props.iconClassName;

    var other = _objectWithoutProperties(_props, ['style', 'children', 'menuItems', 'closeOnMenuItemClick', 'iconStyle', 'iconClassName']);

    var styles = this.getStyles();

    return React.createElement(
      'div',
      _extends({}, other, { style: this.mergeAndPrefix(styles.root, this.props.style) }),
      React.createElement(
        'div',
        { onClick: this._onControlClick },
        React.createElement(FontIcon, {
          className: iconClassName,
          style: iconStyle }),
        this.props.children
      ),
      React.createElement(Menu, {
        ref: 'menuItems',
        style: this.mergeAndPrefix(styles.menu),
        menuItems: menuItems,
        menuItemStyle: styles.menuItem,
        hideable: true,
        visible: this.state.open,
        onItemClick: this._onMenuItemClick })
    );
  },

  _onControlClick: function _onControlClick(e) {
    this.setState({ open: !this.state.open });
  },

  _onMenuItemClick: function _onMenuItemClick(e, key, payload) {
    if (this.props.onChange) this.props.onChange(e, key, payload);

    if (this.props.closeOnMenuItemClick) {
      this.setState({ open: false });
    }
  }
});

module.exports = DropDownIcon;