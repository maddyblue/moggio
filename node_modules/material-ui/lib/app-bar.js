'use strict';

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Typography = require('./styles/typography');
var IconButton = require('./icon-button');
var NavigationMenu = require('./svg-icons/navigation-menu');
var Paper = require('./paper');

var AppBar = React.createClass({
  displayName: 'AppBar',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    onLeftIconButtonTouchTap: React.PropTypes.func,
    onRightIconButtonTouchTap: React.PropTypes.func,
    showMenuIconButton: React.PropTypes.bool,
    iconClassNameLeft: React.PropTypes.string,
    iconClassNameRight: React.PropTypes.string,
    iconElementLeft: React.PropTypes.element,
    iconElementRight: React.PropTypes.element,
    iconStyleRight: React.PropTypes.object,
    title: React.PropTypes.node,
    zDepth: React.PropTypes.number },

  getDefaultProps: function getDefaultProps() {
    return {
      showMenuIconButton: true,
      title: '',
      zDepth: 1
    };
  },

  componentDidMount: function componentDidMount() {
    if (process.env.NODE_ENV !== 'production' && (this.props.iconElementLeft && this.props.iconClassNameLeft)) {
      var warning = 'Properties iconClassNameLeft and iconElementLeft cannot be simultaneously ' + 'defined. Please use one or the other.';
      console.warn(warning);
    }
  },

  getStyles: function getStyles() {
    var spacing = this.context.muiTheme.spacing;
    var themeVariables = this.context.muiTheme.component.appBar;
    var iconButtonSize = this.context.muiTheme.component.button.iconButtonSize;
    var styles = {
      root: {
        zIndex: 5,
        width: '100%',
        minHeight: themeVariables.height,
        backgroundColor: themeVariables.color,
        paddingLeft: spacing.desktopGutter,
        paddingRight: spacing.desktopGutter
      },
      title: {
        float: 'left',
        margin: 0,
        paddingTop: 0,
        letterSpacing: 0,
        fontSize: '24px',
        fontWeight: Typography.fontWeightNormal,
        color: themeVariables.textColor,
        lineHeight: themeVariables.height + 'px'
      },
      iconButton: {
        style: {
          marginTop: (themeVariables.height - iconButtonSize) / 2,
          float: 'left',
          marginRight: 8,
          marginLeft: -16
        },
        iconStyle: {
          fill: themeVariables.textColor,
          color: themeVariables.textColor
        }
      }
    };
    return styles;
  },

  render: function render() {
    var styles = this.getStyles();

    var title, menuElementLeft, menuElementRight;
    var iconRightStyle = this.mergeAndPrefix(styles.iconButton.style, {
      float: 'right',
      marginRight: -16,
      marginLeft: 8
    }, this.props.iconStyleRight);

    if (this.props.title) {
      // If the title is a string, wrap in an h1 tag.
      // If not, just use it as a node.
      title = Object.prototype.toString.call(this.props.title) === '[object String]' ? React.createElement(
        'h1',
        { style: this.mergeAndPrefix(styles.title) },
        this.props.title
      ) : this.props.title;
    }

    if (this.props.showMenuIconButton) {
      if (this.props.iconElementLeft) {
        menuElementLeft = React.createElement(
          'div',
          { style: styles.iconButton.style },
          this.props.iconElementLeft
        );
      } else {
        var child = this.props.iconClassNameLeft ? '' : React.createElement(NavigationMenu, { style: this.mergeAndPrefix(styles.iconButton.iconStyle) });
        menuElementLeft = React.createElement(
          IconButton,
          {
            style: this.mergeAndPrefix(styles.iconButton.style),
            iconStyle: this.mergeAndPrefix(styles.iconButton.iconStyle),
            iconClassName: this.props.iconClassNameLeft,
            onTouchTap: this._onLeftIconButtonTouchTap },
          child
        );
      }

      if (this.props.iconElementRight) {
        menuElementRight = React.createElement(
          'div',
          { style: iconRightStyle },
          this.props.iconElementRight
        );
      } else if (this.props.iconClassNameRight) {
        menuElementRight = React.createElement(IconButton, {
          style: iconRightStyle,
          iconStyle: this.mergeAndPrefix(styles.iconButton.iconStyle),
          iconClassName: this.props.iconClassNameRight,
          onTouchTap: this._onRightIconButtonTouchTap });
      }
    }

    return React.createElement(
      Paper,
      {
        rounded: false,
        className: this.props.className,
        style: this.mergeAndPrefix(styles.root, this.props.style),
        zDepth: this.props.zDepth },
      menuElementLeft,
      title,
      menuElementRight
    );
  },

  _onLeftIconButtonTouchTap: function _onLeftIconButtonTouchTap(e) {
    if (this.props.onLeftIconButtonTouchTap) this.props.onLeftIconButtonTouchTap(e);
  },

  _onRightIconButtonTouchTap: function _onRightIconButtonTouchTap(e) {
    if (this.props.onRightIconButtonTouchTap) this.props.onRightIconButtonTouchTap(e);
  }

});

module.exports = AppBar;