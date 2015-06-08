'use strict';

var _extends = Object.assign || function (target) { for (var i = 1; i < arguments.length; i++) { var source = arguments[i]; for (var key in source) { if (Object.prototype.hasOwnProperty.call(source, key)) { target[key] = source[key]; } } } return target; };

function _objectWithoutProperties(obj, keys) { var target = {}; for (var i in obj) { if (keys.indexOf(i) >= 0) continue; if (!Object.prototype.hasOwnProperty.call(obj, i)) continue; target[i] = obj[i]; } return target; }

var React = require('react');
var StylePropable = require('./mixins/style-propable');
var Transitions = require('./styles/transitions');
var Colors = require('./styles/colors');

var Tooltip = React.createClass({
  displayName: 'Tooltip',

  mixins: [StylePropable],

  contextTypes: {
    muiTheme: React.PropTypes.object
  },

  propTypes: {
    className: React.PropTypes.string,
    label: React.PropTypes.string.isRequired,
    show: React.PropTypes.bool,
    touch: React.PropTypes.bool
  },

  componentDidMount: function componentDidMount() {
    this._setRippleSize();
  },

  componentDidUpdate: function componentDidUpdate(prevProps, prevState) {
    this._setRippleSize();
  },

  getStyles: function getStyles() {
    var styles = {
      root: {
        position: 'absolute',
        fontFamily: this.context.muiTheme.contentFontFamily,
        fontSize: '10px',
        lineHeight: '22px',
        padding: '0 8px',
        color: Colors.white,
        overflow: 'hidden',
        top: -10000,
        borderRadius: 2,
        userSelect: 'none',
        opacity: 0,
        transition: Transitions.easeOut('0ms', 'top', '450ms') + ',' + Transitions.easeOut('450ms', 'transform', '0ms') + ',' + Transitions.easeOut('450ms', 'opacity', '0ms')
      },
      label: {
        position: 'relative',
        whiteSpace: 'nowrap'
      },
      ripple: {
        position: 'absolute',
        left: '50%',
        top: 0,
        transform: 'translate(-50%, -50%)',
        borderRadius: '50%',
        backgroundColor: 'transparent',
        transition: Transitions.easeOut('0ms', 'width', '450ms') + ',' + Transitions.easeOut('0ms', 'height', '450ms') + ',' + Transitions.easeOut('450ms', 'backgroundColor', '0ms')
      },
      rootWhenShown: {
        top: -16,
        opacity: 1,
        transform: 'translate3d(0px, 16px, 0px)',
        transition: Transitions.easeOut('0ms', 'top', '0ms') + ',' + Transitions.easeOut('450ms', 'transform', '0ms') + ',' + Transitions.easeOut('450ms', 'opacity', '0ms') },
      rootWhenTouched: {
        fontSize: '14px',
        lineHeight: '44px',
        padding: '0 16px'
      },
      rippleWhenShown: {
        backgroundColor: Colors.grey600,
        transition: Transitions.easeOut('450ms', 'width', '0ms') + ',' + Transitions.easeOut('450ms', 'height', '0ms') + ',' + Transitions.easeOut('450ms', 'backgroundColor', '0ms')
      }
    };
    return styles;
  },

  render: function render() {
    var _props = this.props;
    var label = _props.label;

    var other = _objectWithoutProperties(_props, ['label']);

    var styles = this.getStyles();
    return React.createElement(
      'div',
      _extends({}, other, {
        style: this.mergeAndPrefix(styles.root, this.props.show && styles.rootWhenShown, this.props.touch && styles.rootWhenTouched, this.props.style) }),
      React.createElement('div', {
        ref: 'ripple',
        style: this.mergeAndPrefix(styles.ripple, this.props.show && styles.rippleWhenShown) }),
      React.createElement(
        'span',
        { style: this.mergeAndPrefix(styles.label) },
        this.props.label
      )
    );
  },

  _setRippleSize: function _setRippleSize() {
    var ripple = React.findDOMNode(this.refs.ripple);
    var tooltip = window.getComputedStyle(React.findDOMNode(this));
    var tooltipWidth = parseInt(tooltip.getPropertyValue('width'), 10);
    var tooltipHeight = parseInt(tooltip.getPropertyValue('height'), 10);

    var rippleDiameter = Math.sqrt(Math.pow(tooltipHeight, 2) + Math.pow(tooltipWidth / 2, 2)) * 2;

    if (this.props.show) {
      ripple.style.height = rippleDiameter + 'px';
      ripple.style.width = rippleDiameter + 'px';
    } else {
      ripple.style.width = '0px';
      ripple.style.height = '0px';
    }
  }

});

module.exports = Tooltip;