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
		if (this.props.isqueue) {
			POST('/api/cmd/play_idx?idx=' + this.props.idx);
		} else {
			var params = mkcmd([
				'clear',
				'add-' + this.props.id.UID
			]);
			POST('/api/queue/change', params, function() {
				POST('/api/cmd/play');
			});
		}
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
	over: function() {
		this.setState({over: true});
	},
	out: function() {
		this.setState({over: false});
	},
	dequeue: function() {
		var params = mkcmd([
			'rem-' + this.props.idx
		]);
		POST('/api/queue/change', params);
	},
	append: function() {
		var params = mkcmd([
			'add-' + this.props.id.UID
		]);
		POST('/api/queue/change', params);
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
		var control;
		var track;
		var icon = "fa fa-border fa-lg clickable ";
		if (this.state.over) {
			if (this.props.isqueue) {
				control = React.createElement("i", {className: icon + "fa-times", onClick: this.dequeue});
			} else {
				control = React.createElement("i", {className: icon + "fa-plus", onClick: this.append});
			}
			track = React.createElement("i", {className: icon + "fa-play", onClick: this.play});
		} else {
			track = info.Track || '';
			if (this.props.useIdxAsNum) {
				track = this.props.idx + 1;
			}
		}
		return (
			React.createElement("tr", {onMouseEnter: this.over, onMouseLeave: this.out}, 
				React.createElement("td", {className: "control"}, track), 
				React.createElement("td", null, info.Title), 
				React.createElement("td", {className: "control"}, control), 
				React.createElement("td", null, React.createElement(Time, {time: info.Time})), 
				React.createElement("td", null, React.createElement(Link, {to: "artist", params: info}, info.Artist)), 
				React.createElement("td", null, React.createElement(Link, {to: "album", params: info}, info.Album))
			)
		);
	}
});

var Tracks = React.createClass({displayName: "Tracks",
	getInitialState: function() {
		return {
			sort: 'Title',
			asc: true,
		};
	},
	mkparams: function() {
		return _.map(this.props.tracks, function(t) {
			return 'add-' + t.ID.UID;
		});
	},
	play: function() {
		var params = this.mkparams();
		params.unshift('clear');
		POST('/api/queue/change', mkcmd(params), function() {
			POST('/api/cmd/play');
		});
	},
	add: function() {
		var params = this.mkparams();
		POST('/api/queue/change', mkcmd(params));
	},
	sort: function(field) {
		return function() {
			if (this.state.sort == field) {
				this.setState({asc: !this.state.asc});
			} else {
				this.setState({sort: field});
			}
		}.bind(this);
	},
	sortClass: function(field) {
		if (this.props.isqueue) {
			return '';
		}
		var name = 'clickable ';
		if (this.state.sort == field) {
			name += this.state.asc ? 'sort-asc' : 'sort-desc';
		}
		return name;
	},
	render: function() {
		var sorted = this.props.tracks;
		if (!this.props.isqueue) {
			sorted = _.sortBy(this.props.tracks, function(v) {
				if (!v.Info) {
					return v.ID.UID;
				}
				var d = v.Info[this.state.sort];
				if (_.isString(d)) {
					d = d.toLocaleLowerCase();
				}
				return d;
			}.bind(this));
			if (!this.state.asc) {
				sorted.reverse();
			}
		}
		var tracks = _.map(sorted, function(t, idx) {
			return React.createElement(Track, {key: idx + '-' + t.ID.UID, id: t.ID, info: t.Info, idx: idx, isqueue: this.props.isqueue, useIdxAsNum: this.props.useIdxAsNum});
		}.bind(this));
		var queue;
		if (!this.props.isqueue) {
			queue = (
				React.createElement("div", null, 
					React.createElement("button", {onClick: this.play}, "play"), 
					React.createElement("button", {onClick: this.add}, "add")
				)
			);
		};
		var track = this.props.isqueue ? React.createElement("th", null) : React.createElement("th", {className: this.sortClass('Track'), onClick: this.sort('Track')}, "#");
		return (
			React.createElement("div", null, 
				queue, 
				React.createElement("table", {className: "u-full-width tracks"}, 
					React.createElement("thead", null, 
						React.createElement("tr", null, 
							track, 
							React.createElement("th", {className: this.sortClass('Title'), onClick: this.sort('Title')}, "Name"), 
							React.createElement("th", null), 
							React.createElement("th", {className: this.sortClass('Time'), onClick: this.sort('Time')}, React.createElement("i", {className: "fa fa-clock-o"})), 
							React.createElement("th", {className: this.sortClass('Artist'), onClick: this.sort('Artist')}, "Artist"), 
							React.createElement("th", {className: this.sortClass('Album'), onClick: this.sort('Album')}, "Album")
						)
					), 
					React.createElement("tbody", null, tracks)
				)

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
		return React.createElement(Tracks, {tracks: this.state.Tracks});
	}
});

function searchClass(field) {
	return React.createClass({
		mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
		render: function() {
			if (!Stores.tracks.data) {
				return null;
			}
			var tracks = [];
			var prop = this.props.params[field];
			_.each(Stores.tracks.data.Tracks, function(val) {
				if (val.Info[field] == prop) {
					tracks.push(val);
				}
			});
			return React.createElement(Tracks, {tracks: tracks});
		}
	});
}

var Artist = searchClass('Artist');
var Album = searchClass('Album');
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
	refresh: function() {
		POST('/api/protocol/refresh', {
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
					React.createElement("button", {onClick: this.remove}, "remove"), 
					React.createElement("button", {onClick: this.refresh}, "refresh")
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

var Queue = React.createClass({displayName: "Queue",
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	clear: function() {
		var params = mkcmd([
			'clear',
		]);
		POST('/api/queue/change', params);
	},
	save: function() {
		var name = prompt("Playlist name:");
		if (!name) {
			return;
		}
		if (this.state.Playlists[name]) {
			if (!window.confirm("Overwrite existing playlist?")) {
				return;
			}
		}
		var params = _.map(this.state.Queue, function(t) {
			return 'add-' + t.UID;
		});
		params.unshift('clear');
		POST('/api/playlist/change/' + name, mkcmd(params));
	},
	render: function() {
		var q = _.map(this.state.Queue, function(val) {
			return {
				ID: val
			};
		});
		return (
			React.createElement("div", null, 
				React.createElement("h4", null, "Queue"), 
				React.createElement("button", {onClick: this.clear}, "clear"), 
				React.createElement("button", {onClick: this.save}, "save"), 
				React.createElement(Tracks, {tracks: q, isqueue: true})
			)
		);
	}
});

var Playlist = React.createClass({displayName: "Playlist",
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {
			Playlists: {}
		};
	},
	clear: function() {
		if (!confirm("Delete playlist?")) {
			return;
		}
		var params = mkcmd([
			'clear',
		]);
		POST('/api/playlist/change/' + this.props.params.Playlist, params);
	},
	render: function() {
		var q = _.map(this.state.Playlists[this.props.params.Playlist], function(val) {
			return {
				ID: val
			};
		});
		return (
			React.createElement("div", null, 
				React.createElement("h4", null, this.props.params.Playlist), 
				React.createElement("button", {onClick: this.clear}, "delete playlist"), 
				React.createElement(Tracks, {tracks: q, useIdxAsNum: true})
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
var Redirect = Router.Redirect;

var App = React.createClass({displayName: "App",
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	componentDidMount: function() {
		this.startWS();
	},
	getInitialState: function() {
		return {};
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
		var playlists = _.map(this.state.Playlists, function(_, key) {
			return React.createElement("li", {key: key}, React.createElement(Link, {to: "playlist", params: {Playlist: key}}, key));
		});
		return (
			React.createElement("div", null, 
				React.createElement("header", null, 
					React.createElement("ul", null, 
						React.createElement("li", null, React.createElement(Link, {to: "app"}, "Music")), 
						React.createElement("li", null, React.createElement(Link, {to: "protocols"}, "Sources")), 
						React.createElement("li", null, React.createElement(Link, {to: "queue"}, "Queue"))
					), 
					React.createElement("h4", null, "Playlists"), 
					React.createElement("ul", null, playlists)
				), 
				React.createElement("main", null, 
					React.createElement(RouteHandler, React.__spread({},  this.props))
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
			var info = Lookup(this.state.Song);
			var song = this.state.Song.UID;
			if (info) {
				info = info.Info;
				var album;
				if (info.Album) {
					album = React.createElement("span", null, "- ", React.createElement(Link, {to: "album", params: info}, info.Album));
				}
				var artist;
				if (info.Artist) {
					artist = React.createElement("span", null, "- ", React.createElement(Link, {to: "artist", params: info}, info.Artist));
				}
				song = (
					React.createElement("span", null, 
						info.Title, 
						album, 
						artist
					)
				);
			}
			status = (
				React.createElement("span", null, 
					React.createElement("span", null, 
						React.createElement(Time, {time: this.state.Elapsed}), " /", 
						React.createElement(Time, {time: this.state.Time})
					), 
					song
				)
			);
		};

		var play = 'fa-stop';
		switch(this.state.State) {
			case 0:
				play = 'fa-pause';
				break;
			case 2:
			default:
				play = 'fa-play';
				break;
		}
		var icon = 'fa fa-fw fa-border fa-2x clickable ';
		var repeat = this.state.Repeat ? 'highlight ' : '';
		var random = this.state.Random ? 'highlight ' : '';
		return (
			React.createElement("div", null, 
				React.createElement("span", null, React.createElement("i", {className: icon + repeat + 'fa-repeat', onClick: this.cmd('repeat')})), 
				React.createElement("span", null, React.createElement("i", {className: icon + 'fa-fast-backward', onClick: this.cmd('prev')})), 
				React.createElement("span", null, React.createElement("i", {className: icon + play, onClick: this.cmd('pause')})), 
				React.createElement("span", null, React.createElement("i", {className: icon + 'fa-fast-forward', onClick: this.cmd('next')})), 
				React.createElement("span", null, React.createElement("i", {className: icon + random + 'fa-random', onClick: this.cmd('random')})), 
				React.createElement("span", null, status)
			)
		);
	}
});

var routes = (
	React.createElement(Route, {name: "app", path: "/", handler: App}, 
		React.createElement(DefaultRoute, {handler: TrackList}), 
		React.createElement(Route, {name: "album", path: "/album/:Album", handler: Album}), 
		React.createElement(Route, {name: "artist", path: "/artist/:Artist", handler: Artist}), 
		React.createElement(Route, {name: "playlist", path: "/playlist/:Playlist", handler: Playlist}), 
		React.createElement(Route, {name: "protocols", handler: Protocols}), 
		React.createElement(Route, {name: "queue", handler: Queue})
	)
);

Router.run(routes, Router.HistoryLocation, function (Handler, state) {
	var params = state.params;
	React.render(React.createElement(Handler, {params: params}), document.getElementById('main'));
});
