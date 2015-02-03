// @flow

var Actions = Reflux.createActions([
	'playlist',
	'protocols',
	'status',
	'tracks',
]);

var Stores = {};

_.each(Actions, function(action, name) {
	Stores[name] = Reflux.createStore({
		init: function() {
			this.listenTo(action, this.update);
		},
		update: function(data) {
			this.data = data;
			this.trigger.apply(this, arguments);
		}
	});
});

function POST(path, params, success) {
	var data;
	if (_.isArray(params)) {
		data = _.map(params, function(v) {
			return encodeURIComponent(v.name) + '=' + encodeURIComponent(v.value);
		}).join('&');
	} else if (_.isObject(params)) {
		data = _.map(params, function(v, k) {
			return encodeURIComponent(k) + '=' + encodeURIComponent(v);
		}).join('&');
	} else {
		data = params;
	}
	var xhr = new XMLHttpRequest();
	xhr.open('POST', path, true);
	xhr.setRequestHeader('Content-type', 'application/x-www-form-urlencoded; charset=UTF-8');
	if (success) {
		xhr.onload = success;
	}
	xhr.send(data);
}

function mkcmd(cmds) {
	return _.map(cmds, function(val) {
		return {
			"name": "c",
			"value": val
		};
	});
}

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

var Time = React.createClass({
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

function mkIcon(name) {
	return 'icon fa fa-border fa-lg clickable ' + name;
}