'use strict';

var React = require('react/addons');
var TabTemplate = require('./tabTemplate');
var InkBar = require('../ink-bar');
var StylePropable = require('../mixins/style-propable.js');
var Colors = require('../styles/colors.js');
var Events = require('../utils/events');

var Tabs = React.createClass({
  displayName: 'Tabs',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    initialSelectedIndex: React.PropTypes.number,
    onActive: React.PropTypes.func,
    tabWidth: React.PropTypes.number
  },

  getInitialState: function getInitialState() {
    var selectedIndex = 0;
    if (this.props.initialSelectedIndex && this.props.initialSelectedIndex < this.props.children.length) {
      selectedIndex = this.props.initialSelectedIndex;
    }
    return {
      selectedIndex: selectedIndex
    };
  },

  getEvenWidth: function getEvenWidth() {
    return parseInt(window.getComputedStyle(React.findDOMNode(this)).getPropertyValue('width'), 10);
  },

  componentDidMount: function componentDidMount() {
    this._updateTabWidth();
    Events.on(window, 'resize', this._updateTabWidth);
  },

  componentWillUnmount: function componentWillUnmount() {
    Events.off(window, 'resize', this._updateTabWidth);
  },

  componentWillReceiveProps: function componentWillReceiveProps(newProps) {
    if (newProps.hasOwnProperty('style')) this._updateTabWidth();
  },

  handleTouchTap: function handleTouchTap(tabIndex, tab) {
    if (this.props.onChange && this.state.selectedIndex !== tabIndex) {
      this.props.onChange(tabIndex, tab);
    }

    this.setState({ selectedIndex: tabIndex });
    //default CB is _onActive. Can be updated in tab.jsx
    if (tab.props.onActive) tab.props.onActive(tab);
  },

  getStyles: function getStyles() {
    var themeVariables = this.context.muiTheme.component.tabs;

    return {
      root: {
        position: 'relative'
      },
      tabItemContainer: {
        margin: '0',
        padding: '0',
        width: '100%',
        height: '48px',
        backgroundColor: themeVariables.backgroundColor,
        whiteSpace: 'nowrap',
        display: 'table'
      }
    };
  },

  render: function render() {
    var styles = this.getStyles();

    var width = this.state.fixedWidth ? 100 / this.props.children.length + '%' : this.props.tabWidth + 'px';

    var left = 'calc(' + width + '*' + this.state.selectedIndex + ')';

    var currentTemplate;
    var tabs = React.Children.map(this.props.children, function (tab, index) {
      if (tab.type.displayName === 'Tab') {
        if (this.state.selectedIndex === index) currentTemplate = tab.props.children;
        return React.addons.cloneWithProps(tab, {
          key: index,
          selected: this.state.selectedIndex === index,
          tabIndex: index,
          width: width,
          handleTouchTap: this.handleTouchTap
        });
      } else {
        var type = tab.type.displayName || tab.type;
        throw 'Tabs only accepts Tab Components as children. Found ' + type + ' as child number ' + (index + 1) + ' of Tabs';
      }
    }, this);

    return React.createElement(
      'div',
      { style: this.mergeAndPrefix(styles.root, this.props.style) },
      React.createElement(
        'div',
        { style: this.mergeAndPrefix(styles.tabItemContainer, this.props.tabItemContainerStyle) },
        tabs
      ),
      React.createElement(InkBar, { left: left, width: width }),
      React.createElement(
        TabTemplate,
        null,
        currentTemplate
      )
    );
  },

  _tabWidthPropIsValid: function _tabWidthPropIsValid() {
    return this.props.tabWidth && this.props.tabWidth * this.props.children.length <= this.getEvenWidth();
  },

  // Validates that the tabWidth can fit all tabs on the tab bar. If not, the
  // tabWidth is recalculated and fixed.
  _updateTabWidth: function _updateTabWidth() {
    if (this._tabWidthPropIsValid()) {
      this.setState({
        fixedWidth: false
      });
    } else {
      this.setState({
        fixedWidth: true
      });
    }
  }

});

module.exports = Tabs;