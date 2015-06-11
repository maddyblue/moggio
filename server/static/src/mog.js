// @flow

var exports = module.exports = {};

var React = require('react');
var Reflux = require('reflux');
var _ = require('underscore');

Actions = exports.Actions = Reflux.createActions([
	'active', // active song
	'error',
	'playlist',
	'protocols',
	'status',
	'tracks',
]);

Stores = exports.Stores = {};

_.each(exports.Actions, function(action, name) {
	exports.Stores[name] = Reflux.createStore({
		init: function() {
			this.listenTo(action, this.update);
		},
		update: function(data) {
			this.data = data;
			this.trigger.apply(this, arguments);
		}
	});
});

var POST = exports.POST = function(path, params, success) {
	var data = new(FormData);
	if (_.isArray(params)) {
		_.each(params, function(v) {
			data.append(v.name, v.value);
		});
	} else if (_.isObject(params)) {
		_.each(params, function(v, k) {
			data.append(k, v);
		});
	} else if (params) {
		data = params;
	}
	var f = fetch(path, {
		method: 'post',
		body: data
	});
	f.then(function(response) {
		if (response.status >= 200 && response.status < 300) {
			return Promise.resolve(response);
		} else {
			return Promise.reject(new Error(response.statusText));
		}
	});
	f.catch(function(err) {
		Actions.error({
			Error: err,
			Time: new Date(),
		});
	});
	if (success) {
		f.then(success);
	}
	return f;
}

exports.mkcmd = function(cmds) {
	return _.map(cmds, function(val) {
		return {
			"name": "c",
			"value": val
		};
	});
};

document.addEventListener('keydown', function(e) {
	if (document.activeElement != document.body) {
		return;
	}
	var cmd;
	switch (e.keyCode) {
	case 32: // space
		cmd = 'pause';
		break;
	case 37: // left
		cmd = 'prev';
		break;
	case 39: // right
		cmd = 'next';
		break;
	default:
		return;
	}
	POST('/api/cmd/' + cmd);
	e.preventDefault();
});

exports.Icon = React.createClass({
	render: function() {
		var cn = 'material-icons';
		if (this.props.className) {
			cn += ' ' + this.props.className;
		}
		return <i {...this.props} className={cn} >{this.props.name}</i>;
	}
});

exports.Time = React.createClass({
	render: function() {
		var t = this.props.time / 1e9;
		var m = Math.floor(t / 60);
		var s = Math.floor(t % 60);
		if (s < 10) {
			s = "0" + s;
		}
		return <span>{m}:{s}</span>;
	}
});
