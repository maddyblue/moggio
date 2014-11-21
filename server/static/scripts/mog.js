(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
/** @jsx React.DOM */

var TrackListRow = React.createClass({displayName: 'TrackListRow',
	render: function() {
		return (React.createElement("tr", null, React.createElement("td", null, this.props.protocol), React.createElement("td", null, this.props.id)));
	}
});

var Track = React.createClass({displayName: 'Track',
	play: function() {
		var params = {
			"clear": true,
			"add": JSON.stringify(this.props.ID)
		};
		$.get('/api/playlist/change?' + $.param(params))
			.success(function() {
				$.get('/api/cmd/play');
			});
	},
	render: function() {
		return (
			React.createElement("tr", null, 
				React.createElement("td", null, React.createElement("button", {onClick: this.play}, "▶"), " ", this.props.Info.Title), 
				React.createElement("td", null, this.props.Info.Artist), 
				React.createElement("td", null, this.props.Info.Album)
			)
		);
	}
});

var TrackList = React.createClass({displayName: 'TrackList',
	getInitialState: function() {
		return {
			tracks: []
		};
	},
	componentDidMount: function() {
		$.get('/api/list', function(result) {
			this.setState({tracks: result});
		}.bind(this));
	},
	render: function() {
		var tracks = this.state.tracks.map(function (t) {
			return React.createElement(Track, React.__spread({},  t));
		});
		return (
			React.createElement("table", null, 
				React.createElement("thead", null, 
					React.createElement("tr", null, 
						React.createElement("th", null, "Name"), 
						React.createElement("th", null, "Artist"), 
						React.createElement("th", null, "Album")
					)
				), 
				React.createElement("tbody", null, tracks)
			)
		);
	}
});

var ProtocolParam = React.createClass({displayName: 'ProtocolParam',
	getInitialState: function() {
		return {
			value: '',
			changed: false,
		};
	},
	componentWillReceiveProps: function(props) {
		if (this.state.changed) {
			return;
		}
		this.setState({
			value: props.value,
			changed: true,
		});
	},
	paramChange: function(event) {
		this.setState({
			value: event.target.value,
		});
		this.props.change();
	},
	render: function() {
		return (
			React.createElement("li", null, 
				this.props.key, " ", React.createElement("input", {type: "text", onChange: this.paramChange, value: this.state.value || this.props.value, disabled: this.props.disabled ? 'disabled' : ''})
			)
		);
	}
});

var ProtocolOAuth = React.createClass({displayName: 'ProtocolOAuth',
	render: function() {
		var token;
		if (this.props.token) {
			token = React.createElement("div", null, "Connected until ", this.props.token.expiry);
		}
		return (
			React.createElement("li", null, 
				token, 
				React.createElement("a", {href: this.props.url}, "connect")
			)
		);
	}
});

var Protocol = React.createClass({displayName: 'Protocol',
	getInitialState: function() {
		return {
			save: false,
		};
	},
	getDefaultProps: function() {
		return {
			instance: {},
			params: {},
		};
	},
	setSave: function() {
		this.setState({save: true});
	},
	save: function() {
		var params = Object.keys(this.refs).sort();
		params = params.map(function(ref) {
			return {
				name: 'params',
				value: this.refs[ref].state.value,
			};
		}, this);
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		$.get('/api/protocol/add?' + $.param(params))
			.success(function() {
				this.setState({save: false});
			}.bind(this))
			.error(function(result) {
				alert(result.responseText);
			});
	},
	remove: function() {
		$.get('/api/protocol/remove?' + $.param({
			protocol: this.props.protocol,
			key: this.props.name,
		}));
	},
	render: function() {
		var params = [];
		var disabled = !!this.props.name;
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var current = this.props.instance.Params || [];
				return React.createElement(ProtocolParam, {key: param, ref: idx, value: current[idx], change: this.setSave, disabled: disabled});
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(React.createElement(ProtocolOAuth, {key: 'oauth-' + this.props.key, url: this.props.params.OAuthURL, token: this.props.instance.OAuthToken, disabled: disabled}));
		}
		var save;
		if (this.state.save) {
			save = React.createElement("button", {onClick: this.save}, "save");
		}
		var title;
		if (this.props.name) {
			title = React.createElement("h3", null, this.props.protocol, ": ", this.props.name, 
					React.createElement("small", null, React.createElement("button", {onClick: this.remove}, "remove"))
				);
		}
		return React.createElement("div", null, 
				title, 
				React.createElement("ul", null, params), 
				save
			);
	}
});

var Protocols = React.createClass({displayName: 'Protocols',
	getInitialState: function() {
		return {
			available: {},
			current: {},
			selected: 'file',
		};
	},
	componentDidMount: function() {
		$.get('/api/protocol/get', function(result) {
			this.setState({available: result});
		}.bind(this));
		$.get('/api/protocol/list', function(result) {
			this.setState({current: result});
		}.bind(this));
	},
	handleChange: function(event) {
		this.setState({selected: event.target.value});
	},
	render: function() {
		var keys = Object.keys(this.state.available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return React.createElement("option", {key: protocol}, protocol);
		}.bind(this));
		var protocols = [];
		_.each(this.state.current, function(instances, protocol) {
			_.each(instances, function(inst, key) {
				protocols.push(React.createElement(Protocol, {key: 'current-' + protocol + '-' + key, protocol: protocol, params: this.state.available[protocol], instance: inst, name: key}));
			}, this);
		}, this);
		var selected;
		if (this.state.selected) {
			selected = React.createElement(Protocol, {key: 'selected-' + this.state.selected, protocol: this.state.selected, params: this.state.available[this.state.selected]});
		}
		return React.createElement("div", null, 
			React.createElement("h2", null, "New Protocol"), 
			React.createElement("select", {onChange: this.handleChange, value: this.state.selected}, options), 
			selected, 
			React.createElement("h2", null, "Existing Protocols"), 
			protocols
		);
	}
});

var routes = {};

var Link = React.createClass({displayName: 'Link',
	componentDidMount: function() {
		routes[this.props.href] = this.props.handler;
		if (this.props.index) {
			routes['/'] = this.props.handler;
		}
	},
	click: function(event) {
		history.pushState(null, this.props.Name, this.props.href);
		router();
		event.preventDefault();
	},
	render: function() {
		return React.createElement("li", null, React.createElement("a", {href: this.props.href, onClick: this.click}, this.props.name))
	}
});

var Navigation = React.createClass({displayName: 'Navigation',
	render: function() {
		return (
			React.createElement("ul", null, 
				React.createElement(Link, {href: "/list", name: "List", handler: TrackList, index: true}), 
				React.createElement(Link, {href: "/protocols", name: "Protocols", handler: Protocols})
			)
		);
	}
});

React.renderComponent(React.createElement(Navigation, null), document.getElementById('navigation'));

function router() {
	var component = routes[window.location.pathname];
	if (!component) {
		alert('unknown route');
	} else {
		React.renderComponent(component(), document.getElementById('main'));
	}
}
router();

var songCache = {};
function getSong(cached, id, cb) {
	var lookup = songCache[id];
	if (lookup) {
		if (lookup != cached) {
			cb({cache: lookup});
		}
		return;
	}
	$.get('/api/song/info?song=' + encodeURIComponent(JSON.stringify(id)))
		.success(function(data) {
			songCache[id] = data[0].Title;
			if (songCache[id] != cached) {
				cb({cache: songCache[id]});
			}
		});
}

var Player = React.createClass({displayName: 'Player',
	getInitialState: function() {
		return {};
	},
	componentDidUpdate: function(props, state) {
		if (state.status && state.status.Song && state.status.Song.ID) {
			getSong(this.state.cache, state.status.Song, this.setState.bind(this));
		}
	},
	startWS: function() {
		console.log('open ws');
		var ws = new WebSocket('ws://' + window.location.host + '/ws/');
		ws.onmessage = function(e) {
			this.setState({status: JSON.parse(e.data)});
		}.bind(this);
		ws.onclose = function() {
			setTimeout(this.startWS, 1000);
		}.bind(this);
	},
	componentDidMount: function() {
		this.startWS();
	},
	cmd: function(cmd) {
		return function() {
			$.get('/api/cmd/' + cmd)
				.error(function(err) {
					console.log(err.responseText);
				});
		};
	},
	render: function() {
		var player = (
			React.createElement("div", null, 
				React.createElement("button", {onClick: this.cmd('prev')}, "prev"), 
				React.createElement("button", {onClick: this.cmd('pause')}, "play/pause"), 
				React.createElement("button", {onClick: this.cmd('next')}, "next")
			)
		);
		var status;
		if (!this.state.status) {
			status = React.createElement("div", null, "unknown");
		} else {
			status = (
				React.createElement("ul", null, 
					React.createElement("li", null, "cache: ", this.state.cache), 
					React.createElement("li", null, "pl: ", this.state.status.Playlist), 
					React.createElement("li", null, "state: ", this.state.status.State), 
					React.createElement("li", null, "song: ", this.state.status.Song), 
					React.createElement("li", null, "elapsed: ", this.state.status.Elapsed), 
					React.createElement("li", null, "time: ", this.state.status.Time)
				)
			);
		};
		return React.createElement("div", null, player, status);
	}
});

React.renderComponent(React.createElement(Player, null), document.getElementById('player'));
},{}]},{},[1]);
