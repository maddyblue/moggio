'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react/addons');
var StylePropable = require('./mixins/style-propable');
var Colors = require('./styles/colors');
var Typography = require('./styles/typography');

var SvgIcon = React.createClass({
  displayName: 'SvgIcon',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    icon: React.PropTypes.element,
    backgroundColor: React.PropTypes.string,
    color: React.PropTypes.string,
    src: React.PropTypes.string
  },

  getDefaultProps: function getDefaultProps() {
    return {
      backgroundColor: Colors.grey400,
      color: Colors.white
    };
  },

  render: function render() {
    var _props = this.props;
    var icon = _props.icon;
    var backgroundColor = _props.backgroundColor;
    var color = _props.color;
    var src = _props.src;
    var style = _props.style;

    var other = _objectWithoutProperties(_props, ['icon', 'backgroundColor', 'color', 'src', 'style']);

    var styles = {
      root: {
        height: src ? 38 : 40,
        width: src ? 38 : 40,
        userSelect: 'none',
        backgroundColor: backgroundColor,
        borderRadius: '50%',
        border: src ? 'solid 1px' : 'none',
        borderColor: this.context.muiTheme.palette.borderColor,
        display: 'inline-block',

        //Needed for letter avatars
        textAlign: 'center',
        lineHeight: '40px',
        fontSize: 24,
        color: color
      },

      iconStyles: {
        margin: 8
      }
    };

    var mergedRootStyles = this.mergeAndPrefix(styles.root, style);
    var mergedIconStyles = icon ? this.mergeStyles(styles.iconStyles, icon.props.style) : null;

    var iconElement = icon ? React.cloneElement(icon, {
      color: color,
      style: mergedIconStyles
    }) : null;

    return src ? React.createElement('img', _extends({}, other, { src: src, style: mergedRootStyles })) : React.createElement(
      'div',
      _extends({}, other, { style: mergedRootStyles }),
      iconElement,
      this.props.children
    );
  }
});

module.exports = SvgIcon;