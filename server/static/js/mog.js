// @flow

var Actions = Reflux.createActions([
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

var Time = React.createClass({displayName: "Time",
	render: function() {
		var t = this.props.time / 1e9;
		var m = Math.floor(t / 60);
		var s = Math.floor(t % 60);
		if (s < 10) {
			s = "0" + s;
		}
		console.log(t, m, s);
		return React.createElement("span", null, m, ":", s);
	}
});
// @flow

var TrackListRow = React.createClass({displayName: "TrackListRow",
	render: function() {
		return (React.createElement("tr", null, React.createElement("td", null, this.props.protocol), React.createElement("td", null, this.props.id)));
	}
});

var Track = React.createClass({displayName: "Track",
	play: function() {
		var params = {
			"clear": true,
			"add": JSON.stringify(this.props.ID)
		};
		POST('/api/playlist/change', params, function() {
				POST('/api/cmd/play');
			});
	},
	render: function() {
		return (
			React.createElement("tr", null, 
				React.createElement("td", null, React.createElement("button", {className: "btn btn-default btn-sm", onClick: this.play}, "▶"), " ", this.props.Info.Title), 
				React.createElement("td", null, React.createElement(Time, {time: this.props.Info.Time})), 
				React.createElement("td", null, this.props.Info.Artist), 
				React.createElement("td", null, this.props.Info.Album)
			)
		);
	}
});

var TrackList = React.createClass({displayName: "TrackList",
	mixins: [Reflux.listenTo(Stores.tracks, 'setTracks')],
	getInitialState: function() {
		return {
			tracks: Stores.tracks.data
		};
	},
	setTracks: function(tracks) {
		var sc = {};
		tracks.forEach(function(t) {
			var uid = t.ID.Protocol + "|" + t.ID.Key + "|" + t.ID.ID;
			sc[uid] = t;
		});
		this.setState({tracks: sc});
	},
	render: function() {
		var tracks = _.map(this.state.tracks, (function (t, key) {
			return React.createElement(Track, React.__spread({key: key},  t));
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

var routes = {};

var Link = React.createClass({displayName: "Link",
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

var Navigation = React.createClass({displayName: "Navigation",
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
		}.bind(this);
		ws.onclose = function() {
			setTimeout(this.startWS, 1000);
		}.bind(this);
	},
	render: function() {
		return (
			React.createElement("ul", {className: "nav navbar-nav"}, 
				React.createElement(Link, {href: "/list", name: "List", handler: TrackList, index: true}), 
				React.createElement(Link, {href: "/protocols", name: "Protocols", handler: Protocols})
			)
		);
	}
});

var navigation = React.createElement(Navigation, null);
React.render(navigation, document.getElementById('navbar'));

function router() {
	var component = routes[window.location.pathname];
	if (!component) {
		alert('unknown route');
	} else {
		React.render(React.createElement(component, {key: window.location.pathname}), document.getElementById('main'));
	}
}
router();

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
		if (!this.state.Song) {
			status = React.createElement("span", null, "unknown");
		} else {
			status = (
				React.createElement("span", null, 
					React.createElement("span", null, "pl: ", this.state.Playlist), 
					React.createElement("span", null, "state: ", this.state.State), 
					React.createElement("span", null, "elapsed: ", React.createElement(Time, {time: this.state.Elapsed})), 
					React.createElement("span", null, "time: ", React.createElement(Time, {time: this.state.Time})), 
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

var player = React.createElement(Player, null);
React.render(player, document.getElementById('player'));
