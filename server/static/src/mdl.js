var React = require('react');

var exports = module.exports = {};

function propClasses(prefix, props) {
	var cn = '';
	for (var i in props) {
		if (!props[i]) {
			continue;
		}
		
		cn += ' ' + prefix + i;
	}
	return cn;
}

exports.TextField = React.createClass({
	componentDidMount: function() {
		componentHandler.upgradeDom();
	},
	render: function() {
		var {
			children,
			error,
			floating,
			pattern,
			...others } = this.props;
		var cn = 'mdl-textfield mdl-js-textfield';
		if (floating) {
			cn += ' mdl-textfield--floating-label';
		}
		if (error) {
			error = <span className="mdl-textfield__error">{error}</span>;
		}
		return (
			<div
				className={cn}
				>
				<input className="mdl-textfield__input" pattern={pattern} {...others} />
				<label className="mdl-textfield__label">{children}</label>
				{error}
			</div>
		);
	}
});

exports.Button = React.createClass({
	render: function() {
		var {
			children,
			disabled,
			ripple,
			...others } = this.props;
		var cn = 'mdl-button mdl-js-button';
		if (ripple) {
			cn += ' mdl-js-ripple-effect';
		}
		cn += propClasses('mdl-button--', others);
		return (
			<button
				className={cn}
				disabled={disabled}
				{...others}
				>
				{children}
			</button>
		);
	}
});
