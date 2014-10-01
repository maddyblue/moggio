(function e(t,n,r){function s(o,u){if(!n[o]){if(!t[o]){var a=typeof require=="function"&&require;if(!u&&a)return a(o,!0);if(i)return i(o,!0);var f=new Error("Cannot find module '"+o+"'");throw f.code="MODULE_NOT_FOUND",f}var l=n[o]={exports:{}};t[o][0].call(l.exports,function(e){var n=t[o][1][e];return s(n?n:e)},l,l.exports,e,t,n,r)}return n[o].exports}var i=typeof require=="function"&&require;for(var o=0;o<r.length;o++)s(r[o]);return s})({1:[function(require,module,exports){
/** @jsx React.DOM */

var TrackListRow = React.createClass({displayName: 'TrackListRow',
	render: function() {
		return (React.DOM.tr(null, React.DOM.td(null, this.props.protocol), React.DOM.td(null, this.props.id)));
	}
});

var Track = React.createClass({displayName: 'Track',
	play: function() {
		var params = {
			"clear": true,
			"add": this.props.key
		};
		$.get('/api/playlist/change?' + $.param(params))
			.success(function() {
				$.get('/api/cmd/play');
			});
	},
	render: function() {
		return (
			React.DOM.tr(null, 
				React.DOM.td(null, React.DOM.button({onClick: this.play}, "▶")), 
				React.DOM.td(null, this.props.key)
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
			return Track({key: t});
		});
		return (
			React.DOM.table(null, 
				React.DOM.tbody(null, tracks)
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
			React.DOM.li(null, 
				this.props.key, " ", React.DOM.input({type: "text", onChange: this.paramChange, value: this.state.value})
			)
		);
	}
});

var ProtocolOAuth = React.createClass({displayName: 'ProtocolOAuth',
	render: function() {
		var token;
		if (this.props.token) {
			token = React.DOM.div(null, "Connected until ", this.props.token.Expiry);
		}
		return (
			React.DOM.li(null, 
				token, 
				React.DOM.a({href: this.props.url}, "connect")
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
			current: {},
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
			value: this.props.key,
		});
		$.get('/api/protocol/update?' + $.param(params))
			.success(function() {
				this.setState({save: false});
			}.bind(this))
			.error(function(result) {
				alert(result.responseText);
			});
	},
	render: function() {
		var params = [];
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var current = this.props.current.Params || [];
				return ProtocolParam({key: param, ref: idx, value: current[idx], change: this.setSave});
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(ProtocolOAuth({url: this.props.params.OAuthURL, token: this.props.current.OAuthToken}));
		}
		var save;
		if (this.state.save) {
			save = React.DOM.button({onClick: this.save}, "save");
		}
		return (
			React.DOM.div({key: this.props.key}, 
				React.DOM.h2(null, this.props.key), 
				React.DOM.ul(null, params), 
				save
			)
		);
	}
});

var Protocols = React.createClass({displayName: 'Protocols',
	getInitialState: function() {
		return {
			available: {},
			current: {},
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
	render: function() {
		var keys = Object.keys(this.state.available);
		keys.sort();
		var protocols = keys.map(function(protocol) {
			return Protocol({key: protocol, params: this.state.available[protocol], current: this.state.current[protocol]});
		}.bind(this));
		return React.DOM.div(null, protocols);
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
		return React.DOM.li(null, React.DOM.a({href: this.props.href, onClick: this.click}, this.props.name))
	}
});

var Navigation = React.createClass({displayName: 'Navigation',
	render: function() {
		return (
			React.DOM.ul(null, 
				Link({href: "/list", name: "List", handler: TrackList, index: true}), 
				Link({href: "/protocols", name: "Protocols", handler: Protocols})
			)
		);
	}
});

React.renderComponent(Navigation(null), document.getElementById('navigation'));

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
	$.get('/api/song/info?song=' + encodeURIComponent(id))
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
		if (state.status && state.status.Song) {
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
			React.DOM.div(null, 
				React.DOM.button({onClick: this.cmd('prev')}, "prev"), 
				React.DOM.button({onClick: this.cmd('pause')}, "play/pause"), 
				React.DOM.button({onClick: this.cmd('next')}, "next")
			)
		);
		var status;
		if (!this.state.status) {
			status = React.DOM.div(null, "unknown");
		} else {
			status = (
				React.DOM.ul(null, 
					React.DOM.li(null, "cache: ", this.state.cache), 
					React.DOM.li(null, "pl: ", this.state.status.Playlist), 
					React.DOM.li(null, "state: ", this.state.status.State), 
					React.DOM.li(null, "song: ", this.state.status.Song), 
					React.DOM.li(null, "elapsed: ", this.state.status.Elapsed), 
					React.DOM.li(null, "time: ", this.state.status.Time)
				)
			);
		};
		return React.DOM.div(null, player, status);
	}
});

React.renderComponent(Player(null), document.getElementById('player'));
},{}]},{},[1]);
