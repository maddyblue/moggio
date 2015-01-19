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

// Lookup returns track data for song with given ID. null is returned if no
// song with id found.
function Lookup(id) {
	var t = Stores.tracks.data;
	if (!t || !t.Tracks) {
		return null;
	}
	t = t.Tracks;
	for (var i = 0; i < t.length; i++) {
		var d = t[i];
		if (_.isEqual(d.ID, id)) {
			return d;
		}
	}
	return null;
}

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

var Time = React.createClass({displayName: "Time",
	render: function() {
		var t = this.props.time / 1e9;
		var m = Math.floor(t / 60);
		var s = Math.floor(t % 60);
		if (s < 10) {
			s = "0" + s;
		}
		return React.createElement("span", null, m, ":", s);
	}
});
// @flow

var Track = React.createClass({displayName: "Track",
	mixins: [Reflux.listenTo(Stores.tracks, 'update')],
	play: function() {
		var params = {
			"clear": true,
			"add": JSON.stringify(this.props.id)
		};
		POST('/api/playlist/change', params, function() {
				POST('/api/cmd/play');
			});
	},
	getInitialState: function() {
		if (this.props.info) {
			return {
				info: this.props.info
			};
		}
		var d = Lookup(this.props.id);
		if (d) {
			return {
				info: d.Info
			};
		}
		return {};
	},
	update: function() {
		this.setState(this.getInitialState());
	},
	render: function() {
		var info = this.state.info;
		if (!info) {
			return (
				React.createElement("tr", null, 
					React.createElement("td", null, this.props.id)
				)
			);
		}
		return (
			React.createElement("tr", null, 
				React.createElement("td", null, React.createElement("button", {className: "btn btn-default btn-sm", onClick: this.play}, "▶"), " ", info.Title), 
				React.createElement("td", null, React.createElement(Time, {time: info.Time})), 
				React.createElement("td", null, info.Artist), 
				React.createElement("td", null, info.Album)
			)
		);
	}
});

var TrackList = React.createClass({displayName: "TrackList",
	mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
	getInitialState: function() {
		return Stores.tracks.data || {};
	},
	render: function() {
		var tracks = _.map(this.state.Tracks, (function (t) {
			return React.createElement(Track, {key: t.UID, id: t.ID, info: t.Info});
		}));
		return (
			React.createElement("table", {className: "table"}, 
				React.createElement("thead", null, 
					React.createElement("tr", null, 
						React.createElement("th", null, "Name"), 
						React.createElement("th", null, "Time"), 
						React.createElement("th", null, "Artist"), 
						React.createElement("th", null, "Album")
					)
				), 
				React.createElement("tbody", null, tracks)
			)
		);
	}
});
// @flow

var Protocols = React.createClass({displayName: "Protocols",
	mixins: [Reflux.listenTo(Stores.protocols, 'setState')],
	getInitialState: function() {
		var d = {
			Available: {},
			Current: {},
			Selected: 'file',
		};
		return _.extend(d, Stores.protocols.data);
	},
	handleChange: function(event) {
		this.setState({Selected: event.target.value});
	},
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return React.createElement("option", {key: protocol}, protocol);
		}.bind(this));
		var protocols = [];
		_.each(this.state.Current, function(instances, protocol) {
			_.each(instances, function(inst, key) {
				protocols.push(React.createElement(Protocol, {key: key, protocol: protocol, params: this.state.Available[protocol], instance: inst, name: key}));
			}, this);
		}, this);
		var selected;
		if (this.state.Selected) {
			selected = React.createElement(Protocol, {protocol: this.state.Selected, params: this.state.Available[this.state.Selected]});
		}
		return React.createElement("div", null, 
			React.createElement("h2", null, "New Protocol"), 
			React.createElement("select", {onChange: this.handleChange, value: this.state.Selected}, options), 
			selected, 
			React.createElement("h2", null, "Existing Protocols"), 
			protocols
		);
	}
});

var ProtocolParam = React.createClass({displayName: "ProtocolParam",
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
				this.props.name, " ", React.createElement("input", {type: "text", onChange: this.paramChange, value: this.state.value || this.props.value, disabled: this.props.disabled ? 'disabled' : ''})
			)
		);
	}
});

var ProtocolOAuth = React.createClass({displayName: "ProtocolOAuth",
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

var Protocol = React.createClass({displayName: "Protocol",
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
			var v = this.refs[ref].state.value;
			this.refs[ref].state.value = '';
			return {
				name: 'params',
				value: v,
			};
		}, this);
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		POST('/api/protocol/add', params, function() {
				this.setState({save: false});
			}.bind(this));
	},
	remove: function() {
		POST('/api/protocol/remove', {
			protocol: this.props.protocol,
			key: this.props.name,
		});
	},
	render: function() {
		var params = [];
		var disabled = !!this.props.name;
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var current = this.props.instance.Params || [];
				return React.createElement(ProtocolParam, {key: param, name: param, ref: idx, value: current[idx], change: this.setSave, disabled: disabled});
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(React.createElement(ProtocolOAuth, {key: 'oauth', url: this.props.params.OAuthURL, token: this.props.instance.OAuthToken, disabled: disabled}));
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
// @flow

var Playlist = React.createClass({displayName: "Playlist",
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	render: function() {
		var tracks = _.map(this.state.Queue, (function (id) {
			return React.createElement(Track, {key: id.UID, id: id});
		}));
		return (
			React.createElement("table", {className: "table"}, 
				React.createElement("thead", null, 
					React.createElement("tr", null, 
						React.createElement("th", null, "Name"), 
						React.createElement("th", null, "Time"), 
						React.createElement("th", null, "Artist"), 
						React.createElement("th", null, "Album")
					)
				), 
				React.createElement("tbody", null, tracks)
			)
		);
	}
});
// @flow

var Router = ReactRouter;
var Route = Router.Route;
var NotFoundRoute = Router.NotFoundRoute;
var DefaultRoute = Router.DefaultRoute;
var Link = Router.Link;
var RouteHandler = Router.RouteHandler;

var App = React.createClass({displayName: "App",
	componentDidMount: function() {
		this.startWS();
	},
	startWS: function() {
		var ws = new WebSocket('ws://' + window.location.host + '/ws/');
		ws.onmessage = function(e) {
			var d = JSON.parse(e.data);
			if (Actions[d.Type]) {
				Actions[d.Type](d.Data);
			} else {
				console.log("missing action", d.Type);
			}
		};
		ws.onclose = function() {
			setTimeout(this.startWS, 1000);
		}.bind(this);
	},
	render: function() {
		return (
			React.createElement("div", null, 
				React.createElement("header", null, 
					React.createElement("ul", null, 
						React.createElement("li", null, React.createElement(Link, {to: "app"}, "Music")), 
						React.createElement("li", null, React.createElement(Link, {to: "protocols"}, "Sources")), 
						React.createElement("li", null, React.createElement(Link, {to: "playlist"}, "Playlist"))
					)
				), 
				React.createElement("main", null, 
					React.createElement(RouteHandler, null)
				), 
				React.createElement("footer", null, 
					React.createElement(Player, null)
				)
			)
		);
	}
});

var Player = React.createClass({displayName: "Player",
	mixins: [Reflux.listenTo(Stores.status, 'setState')],
	cmd: function(cmd) {
		return function() {
			POST('/api/cmd/' + cmd);
		};
	},
	getInitialState: function() {
		return {};
	},
	render: function() {
		var status;
		if (this.state.Song && this.state.Song.ID) {
			status = (
				React.createElement("span", null, 
					React.createElement("span", null, 
						React.createElement(Time, {time: this.state.Elapsed}), " /", 
						React.createElement(Time, {time: this.state.Time})
					), 
					React.createElement("span", null, "song: ", this.state.Song)
				)
			);
		};

		var play;
		switch(this.state.State) {
			case 0:
				play = '▐▐';
				break;
			case 2:
			default:
				play = '\u25b6';
				break;
		}
		return (
			React.createElement("div", null, 
				React.createElement("span", null, React.createElement("button", {onClick: this.cmd('prev')}, "⇤")), 
				React.createElement("span", null, React.createElement("button", {onClick: this.cmd('pause')}, play)), 
				React.createElement("span", null, React.createElement("button", {onClick: this.cmd('next')}, "⇥")), 
				React.createElement("span", null, status)
			)
		);
	}
});

var routes = (
	React.createElement(Route, {name: "app", path: "/", handler: App}, 
		React.createElement(DefaultRoute, {handler: TrackList}), 
		React.createElement(Route, {name: "protocols", handler: Protocols}), 
		React.createElement(Route, {name: "playlist", handler: Playlist})
	)
);

Router.run(routes, Router.HistoryLocation, function (Handler) {
	React.render(React.createElement(Handler, null), document.getElementById('main'));
});
