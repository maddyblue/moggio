'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var CssEvent = require('../utils/css-event');
var Dom = require('../utils/dom');
var KeyLine = require('../utils/key-line');
var StylePropable = require('../mixins/style-propable');
var Transitions = require('../styles/transitions');
var ClickAwayable = require('../mixins/click-awayable');
var Paper = require('../paper');
var MenuItem = require('./menu-item');
var LinkMenuItem = require('./link-menu-item');
var SubheaderMenuItem = require('./subheader-menu-item');

/***********************
* Nested Menu Component
***********************/
var NestedMenuItem = React.createClass({
  displayName: 'NestedMenuItem',

  mixins: [ClickAwayable, StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    index: React.PropTypes.number.isRequired,
    text: React.PropTypes.string,
    menuItems: React.PropTypes.array.isRequired,
    zDepth: React.PropTypes.number,
    disabled: React.PropTypes.bool,
    onItemClick: React.PropTypes.func,
    onItemTap: React.PropTypes.func,
    menuItemStyle: React.PropTypes.object },

  getDefaultProps: function getDefaultProps() {
    return {
      disabled: false
    };
  },

  getInitialState: function getInitialState() {
    return { open: false };
  },

  componentClickAway: function componentClickAway() {
    this._closeNestedMenu();
  },

  componentDidMount: function componentDidMount() {
    this._positionNestedMenu();
  },

  componentDidUpdate: function componentDidUpdate(prevProps, prevState) {
    this._positionNestedMenu();
  },

  getSpacing: function getSpacing() {
    return this.context.muiTheme.spacing;
  },

  render: function render() {
    var styles = this.mergeAndPrefix({
      position: 'relative'
    }, this.props.style);

    var iconCustomArrowDropRight = {
      marginRight: this.getSpacing().desktopGutterMini * -1,
      color: this.context.muiTheme.component.dropDownMenu.accentColor
    };

    var _props = this.props;
    var index = _props.index;
    var menuItemStyle = _props.menuItemStyle;

    var other = _objectWithoutProperties(_props, ['index', 'menuItemStyle']);

    return React.createElement(
      'div',
      { ref: 'root', style: styles, onMouseEnter: this._openNestedMenu, onMouseLeave: this._closeNestedMenu },
      React.createElement(
        MenuItem,
        {
          index: index,
          style: menuItemStyle,
          disabled: this.props.disabled,
          iconRightStyle: iconCustomArrowDropRight,
          iconRightClassName: 'muidocs-icon-custom-arrow-drop-right',
          onClick: this._onParentItemClick },
        this.props.text
      ),
      React.createElement(Menu, _extends({}, other, {
        ref: 'nestedMenu',
        menuItems: this.props.menuItems,
        onItemClick: this._onMenuItemClick,
        onItemTap: this._onMenuItemTap,
        hideable: true,
        visible: this.state.open,
        zDepth: this.props.zDepth + 1 }))
    );
  },

  _positionNestedMenu: function _positionNestedMenu() {
    var el = React.findDOMNode(this);
    var nestedMenu = React.findDOMNode(this.refs.nestedMenu);

    nestedMenu.style.left = el.offsetWidth + 'px';
  },

  _openNestedMenu: function _openNestedMenu() {
    if (!this.props.disabled) this.setState({ open: true });
  },

  _closeNestedMenu: function _closeNestedMenu() {
    this.setState({ open: false });
  },

  _toggleNestedMenu: function _toggleNestedMenu() {
    if (!this.props.disabled) this.setState({ open: !this.state.open });
  },

  _onParentItemClick: function _onParentItemClick() {
    this._toggleNestedMenu();
  },

  _onMenuItemClick: function _onMenuItemClick(e, index, menuItem) {
    if (this.props.onItemClick) this.props.onItemClick(e, index, menuItem);
    this._closeNestedMenu();
  },

  _onMenuItemTap: function _onMenuItemTap(e, index, menuItem) {
    if (this.props.onItemTap) this.props.onItemTap(e, index, menuItem);
    this._closeNestedMenu();
  }

});

/****************
* Menu Component
****************/
var Menu = React.createClass({
  displayName: 'Menu',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    autoWidth: React.PropTypes.bool,
    onItemTap: React.PropTypes.func,
    onItemClick: React.PropTypes.func,
    onToggle: React.PropTypes.func,
    menuItems: React.PropTypes.array.isRequired,
    selectedIndex: React.PropTypes.number,
    hideable: React.PropTypes.bool,
    visible: React.PropTypes.bool,
    zDepth: React.PropTypes.number,
    menuItemStyle: React.PropTypes.object,
    menuItemStyleSubheader: React.PropTypes.object,
    menuItemStyleLink: React.PropTypes.object,
    menuItemClassName: React.PropTypes.string,
    menuItemClassNameSubheader: React.PropTypes.string,
    menuItemClassNameLink: React.PropTypes.string },

  getInitialState: function getInitialState() {
    return { nestedMenuShown: false };
  },

  getDefaultProps: function getDefaultProps() {
    return {
      autoWidth: true,
      hideable: false,
      visible: true,
      zDepth: 1 };
  },

  componentDidMount: function componentDidMount() {
    var el = React.findDOMNode(this);

    //Set the menu width
    this._setKeyWidth(el);

    //Save the initial menu height for later
    this._initialMenuHeight = el.offsetHeight;

    //Show or Hide the menu according to visibility
    this._renderVisibility();
  },

  componentDidUpdate: function componentDidUpdate(prevProps, prevState) {
    if (this.props.visible !== prevProps.visible) this._renderVisibility();
  },

  getTheme: function getTheme() {
    return this.context.muiTheme.component.menu;
  },

  getSpacing: function getSpacing() {
    return this.context.muiTheme.spacing;
  },

  getStyles: function getStyles() {
    var styles = {
      root: {
        backgroundColor: this.getTheme().containerBackgroundColor,
        paddingTop: this.getSpacing().desktopGutterMini,
        paddingBottom: this.getSpacing().desktopGutterMini,
        transition: Transitions.easeOut(null, 'height')
      },
      subheader: {
        paddingLeft: this.context.muiTheme.component.menuSubheader.padding,
        paddingRight: this.context.muiTheme.component.menuSubheader.padding
      },
      hideable: {
        opacity: this.props.visible ? 1 : 0,
        overflow: 'hidden',
        position: 'absolute',
        top: 0,
        zIndex: 1
      }
    };
    return styles;
  },

  render: function render() {
    var styles = this.getStyles();
    return React.createElement(
      Paper,
      {
        ref: 'paperContainer',
        zDepth: this.props.zDepth,
        style: this.mergeAndPrefix(styles.root, this.props.hideable && styles.hideable, this.props.style) },
      this._getChildren()
    );
  },

  _getChildren: function _getChildren() {
    var children = [],
        menuItem,
        itemComponent,
        isSelected,
        isDisabled;

    var styles = this.getStyles();

    //This array is used to keep track of all nested menu refs
    this._nestedChildren = [];

    for (var i = 0; i < this.props.menuItems.length; i++) {
      menuItem = this.props.menuItems[i];
      isSelected = i === this.props.selectedIndex;
      isDisabled = menuItem.disabled === undefined ? false : menuItem.disabled;

      var icon = menuItem.icon;
      var data = menuItem.data;
      var attribute = menuItem.attribute;
      var number = menuItem.number;
      var toggle = menuItem.toggle;
      var onClick = menuItem.onClick;

      var other = _objectWithoutProperties(menuItem, ['icon', 'data', 'attribute', 'number', 'toggle', 'onClick']);

      switch (menuItem.type) {

        case MenuItem.Types.LINK:
          itemComponent = React.createElement(LinkMenuItem, {
            key: i,
            index: i,
            text: menuItem.text,
            disabled: isDisabled,
            className: this.props.menuItemClassNameLink,
            style: this.props.menuItemStyleLink,
            payload: menuItem.payload,
            target: menuItem.target });
          break;

        case MenuItem.Types.SUBHEADER:
          itemComponent = React.createElement(SubheaderMenuItem, {
            key: i,
            index: i,
            className: this.props.menuItemClassNameSubheader,
            style: this.mergeAndPrefix(styles.subheader),
            firstChild: i === 0,
            text: menuItem.text });
          break;

        case MenuItem.Types.NESTED:
          var _props2 = this.props,
              ref = _props2.ref,
              key = _props2.key,
              index = _props2.index,
              zDepth = _props2.zDepth,
              other = _objectWithoutProperties(_props2, ['ref', 'key', 'index', 'zDepth']);

          itemComponent = React.createElement(NestedMenuItem, _extends({}, other, {
            ref: i,
            key: i,
            index: i,
            text: menuItem.text,
            disabled: isDisabled,
            menuItems: menuItem.items,
            menuItemStyle: this.props.menuItemStyle,
            zDepth: this.props.zDepth,
            onItemClick: this._onNestedItemClick,
            onItemTap: this._onNestedItemClick }));
          this._nestedChildren.push(i);
          break;

        default:
          itemComponent = React.createElement(
            MenuItem,
            _extends({}, other, {
              selected: isSelected,
              key: i,
              index: i,
              icon: menuItem.icon,
              data: menuItem.data,
              className: this.props.menuItemClassName,
              style: this.props.menuItemStyle,
              attribute: menuItem.attribute,
              number: menuItem.number,
              toggle: menuItem.toggle,
              onToggle: this.props.onToggle,
              disabled: isDisabled,
              onClick: this._onItemClick,
              onTouchTap: this._onItemTap }),
            menuItem.text
          );
      }
      children.push(itemComponent);
    }

    return children;
  },

  _setKeyWidth: function _setKeyWidth(el) {
    var menuWidth = this.props.autoWidth ? KeyLine.getIncrementalDim(el.offsetWidth) + 'px' : '100%';

    //Update the menu width
    Dom.withoutTransition(el, function () {
      el.style.width = menuWidth;
    });
  },

  _renderVisibility: function _renderVisibility() {
    var el;

    if (this.props.hideable) {
      el = React.findDOMNode(this);
      var container = React.findDOMNode(this.refs.paperContainer);

      if (this.props.visible) {
        //Open the menu
        el.style.transition = Transitions.easeOut();
        el.style.height = this._initialMenuHeight + 'px';

        //Set the overflow to visible after the animation is done so
        //that other nested menus can be shown
        CssEvent.onTransitionEnd(el, (function () {
          //Make sure the menu is open before setting the overflow.
          //This is to accout for fast clicks
          if (this.props.visible) container.style.overflow = 'visible';
        }).bind(this));
      } else {

        //Close the menu
        el.style.height = '0px';

        //Set the overflow to hidden so that animation works properly
        container.style.overflow = 'hidden';
      }
    }
  },

  _onNestedItemClick: function _onNestedItemClick(e, index, menuItem) {
    if (this.props.onItemClick) this.props.onItemClick(e, index, menuItem);
  },

  _onNestedItemTap: function _onNestedItemTap(e, index, menuItem) {
    if (this.props.onItemTap) this.props.onItemTap(e, index, menuItem);
  },

  _onItemClick: function _onItemClick(e, index) {
    if (this.props.onItemClick) this.props.onItemClick(e, index, this.props.menuItems[index]);
  },

  _onItemTap: function _onItemTap(e, index) {
    if (this.props.onItemTap) this.props.onItemTap(e, index, this.props.menuItems[index]);
  },

  _onItemToggle: function _onItemToggle(e, index, toggled) {
    if (this.props.onItemToggle) this.props.onItemToggle(e, index, this.props.menuItems[index], toggled);
  }

});

module.exports = Menu;