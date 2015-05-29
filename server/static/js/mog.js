// @flow

var Actions = Reflux.createActions([
	'active', // active song
	'error',
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
		alert(err);
	});
	if (success) {
		f.then(success);
	}
	return f;
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

function mkIcon(name) {
	return 'icon fa fa-border fa-lg clickable ' + name;
}
// @flow

var Table = FixedDataTable.Table;
var Column = FixedDataTable.Column;

var Tracks = React.createClass({displayName: "Tracks",
	mixins: [Reflux.listenTo(Stores.active, 'setActive')],
	getDefaultProps: function() {
		return {
			tracks: []
		};
	},
	getInitialState: function() {
		var init = {
			sort: this.props.initSort || 'Title',
			asc: true,
			tracks: [],
			search: '',
		};
		if (this.props.isqueue || this.props.useIdxAsNum) {
			init.sort = 'Track';
		}
		this.update(null, this.props.tracks);
		return init;
	},
	componentWillReceiveProps: function(next) {
		this.update(null, next.tracks);
	},
	setActive: function() {
		this.forceUpdate();
	},
	mkparams: function() {
		return _.map(this.state.tracks, function(t, i) {
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
	playTrack: function(index) {
		return function() {
			if (this.props.isqueue) {
				idx = this.getIdx(index);
				POST('/api/cmd/play_idx?idx=' + idx);
			} else {
				var params = [
					'clear',
					'add-' + this.getter(index).ID.UID
				];
				POST('/api/queue/change', mkcmd(params), function() {
					POST('/api/cmd/play');
				});
			}
		}.bind(this);
	},
	appendTrack: function(index) {
		return function() {
			var params;
			if (this.props.isqueue) {
				var idx = this.getIdx(index);
				params = [
					'rem-' + idx
				];
			} else {
				params = [
					'add-' + this.getter(index).ID.UID
				];
			}
			POST('/api/queue/change', mkcmd(params));
		}.bind(this);
	},
	sort: function(field) {
		return function() {
			if (this.state.sort == field) {
				this.update({asc: !this.state.asc});
			} else {
				this.update({sort: field});
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
	handleResize: function() {
		this.forceUpdate();
	},
	componentDidMount: function() {
		window.addEventListener('resize', this.handleResize);
		this.update();
	},
	componentWillUnmount: function() {
		window.removeEventListener('resize', this.handleResize);
	},
	update: function(obj, next) {
		if (obj) {
			this.setState(obj);
		}
		obj = _.extend({}, this.state, obj);
		var tracks = next || this.props.tracks;
		if (next) {
			_.each(tracks, function(v, i) {
				v.idx = i + 1;
			});
		}
		if (obj.search) {
			var s = obj.search.toLocaleLowerCase().trim();
			tracks = _.filter(tracks, function(v) {
				var t = v.Info.Title + v.Info.Album + v.Info.Artist + v.ID.Protocol;
				t = t.toLocaleLowerCase();
				return t.indexOf(s) > -1;
			});
		}
		var useIdx = (obj.sort == 'Track' && this.props.useIdxAsNum) || this.props.isqueue;
		tracks = _.sortBy(tracks, function(v) {
			return v.Info.Track;
		});
		tracks = _.sortBy(tracks, function(v) {
			if (useIdx) {
				return v.idx;
			}
			var d = v.Info[obj.sort];
			if (obj.sort == "Source") {
				d = v.ID.UID;
			}
			if (_.isString(d)) {
				d = d.toLocaleLowerCase();
			}
			return d;
		}.bind(this));
		if (!obj.asc) {
			tracks.reverse();
		}
		this.setState({tracks: tracks});
	},
	search: function(event) {
		this.update({search: event.target.value});
	},
	getter: function(index) {
		return this.state.tracks[index];
	},
	getIdx: function(index) {
		return this.getter(index).idx - 1;
	},
	timeCellRenderer: function(str, key, data, index) {
		return React.createElement("div", null, React.createElement(Time, {time: data.Info.Time}));
	},
	timeHeader: function() {
		return function() {
			return React.createElement("i", {className: "fa fa-clock-o " + this.sortClass('Time'), onClick: this.sort('Time')});
		}.bind(this);
	},
	mkHeader: function(name, text) {
		if (!text) {
			text = name;
		}
		if (this.props.isqueue) {
			return function() {
				return text;
			};
		}
		return function() {
			return React.createElement("div", {className: this.sortClass(name), onClick: this.sort(name)}, text);
		}.bind(this);
	},
	trackRenderer: function(str, key, data, index) {
		var track = data.Info.Track || '';
		if (this.props.useIdxAsNum) {
			track = data.idx;
		} else if (this.props.noIdx) {
			track = '';
		}
		return (
			React.createElement("div", null, 
				React.createElement("span", {className: "nohover"}, track), 
				React.createElement("span", {className: "hover"}, React.createElement("i", {className: mkIcon('fa-play'), onClick: this.playTrack(index)}))
			)
		);
	},
	titleCellRenderer: function(str, key, data, index) {
		var image;
		if (data.Info.ImageURL) {
			image = React.createElement("img", {className: "track-image", src: data.Info.ImageURL});
		}
		return (
			React.createElement("div", {className: "track-title"}, 
				image, 
				data.Info.Title, 
				React.createElement("span", {className: "hover pull-right"}, React.createElement("i", {className: mkIcon(this.props.isqueue ? 'fa-times' : 'fa-plus'), onClick: this.appendTrack(index)}))
			)
		);
	},
	artistCellRenderer: function(str, key, data, index) {
		return React.createElement("div", null, React.createElement(Link, {to: "artist", params: data.Info}, data.Info.Artist));
	},
	albumCellRenderer: function(str, key, data, index) {
		return React.createElement("div", null, React.createElement(Link, {to: "album", params: data.Info}, data.Info.Album));
	},
	sourceCellRenderer: function(str, key, data, index) {
		return React.createElement("div", {title: data.ID.ID + "|" + data.ID.Key + "|" + data.ID.Protocol}, data.ID.Protocol);
	},
	rowClassNameGetter: function(index) {
		var g = this.getter(index);
		if (g.ID.UID == Stores.active.data) {
			return 'active';
		}
		return null;
	},
	render: function() {
		var height = 0;
		if (this.refs.table) {
			var d = this.refs.table.getDOMNode();
			height = window.innerHeight - d.offsetTop - 82;
		}
		var queue;
		if (!this.props.isqueue) {
			queue = (
				React.createElement("div", null, 
					React.createElement("button", {onClick: this.play}, "play"), 
					" ", 
					React.createElement("button", {onClick: this.add}, "add"), 
					" " + ' ' +
					"(", this.state.tracks.length, " tracks)"
				)
			);
		};
		var track = this.props.isqueue ? React.createElement("th", null) : React.createElement("th", {className: this.sortClass('Track'), onClick: this.sort('Track')}, "#");
		var tableWidth = window.innerWidth - 227;
		return (
			React.createElement("div", null, 
				queue, 
				React.createElement("div", null, React.createElement("input", {type: "search", style: {width: tableWidth - 2}, placeholder: "search", onChange: this.search, value: this.state.search})), 
				React.createElement(Table, {ref: "table", 
					headerHeight: 50, 
					rowHeight: 50, 
					rowGetter: this.getter, 
					rowsCount: this.state.tracks.length, 
					rowClassNameGetter: this.rowClassNameGetter, 
					width: tableWidth, 
					height: height, 
					overflowX: 'hidden'
					}, 
					React.createElement(Column, {
						width: 50, 
						headerRenderer: this.mkHeader('Track', '#'), 
						cellRenderer: this.trackRenderer}
					), 
					React.createElement(Column, {
						width: 200, 
						flexGrow: 3, 
						cellClassName: "nowrap", 
						headerRenderer: this.mkHeader('Title'), 
						cellRenderer: this.titleCellRenderer}
					), 
					React.createElement(Column, {
						width: 50, 
						cellRenderer: this.timeCellRenderer, 
						headerRenderer: this.timeHeader()}
					), 
					React.createElement(Column, {
						width: 100, 
						flexGrow: 1, 
						cellRenderer: this.artistCellRenderer, 
						cellClassName: "nowrap", 
						headerRenderer: this.mkHeader('Artist')}
					), 
					React.createElement(Column, {
						width: 100, 
						flexGrow: 1, 
						cellRenderer: this.albumCellRenderer, 
						cellClassName: "nowrap", 
						headerRenderer: this.mkHeader('Album')}
					), 
					React.createElement(Column, {
						width: 100, 
						cellClassName: "nowrap", 
						cellRenderer: this.sourceCellRenderer, 
						headerRenderer: this.mkHeader('Source')}
					)
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
		return (
			React.createElement("div", null, 
				React.createElement("h4", null, "Music"), 
				React.createElement(Tracks, {tracks: this.state.Tracks})
			)
		);
	}
});

function searchClass(field, sort) {
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
			return React.createElement(Tracks, {tracks: tracks, initSort: sort});
		}
	});
}

var Artist = searchClass('Artist', 'Album');
var Album = searchClass('Album', 'Track');
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
			_.each(instances, function(key) {
				protocols.push(React.createElement(ProtocolRow, {key: key, protocol: protocol, name: key}));
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
			React.createElement("table", null, 
				React.createElement("thead", null, 
					React.createElement("tr", null, 
						React.createElement("th", null, "protocol"), 
						React.createElement("th", null, "name"), 
						React.createElement("th", null, "remove"), 
						React.createElement("th", null, "refresh")
					)
				), 
				React.createElement("tbody", null, 
					protocols
				)
			)
		);
	}
});

var Protocol = React.createClass({displayName: "Protocol",
	getInitialState: function() {
		return {
			params: [],
			save: false,
		};
	},
	save: function() {
		var params = this.state.params.map(function(v) {
			return {
				name: 'params',
				value: v,
			};
		});
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		POST('/api/protocol/add', params, function() {
			this.setState(this.getInitialState());
		}.bind(this));
	},
	render: function() {
		if (!this.props.params) {
			return React.createElement("div", null);
		}
		var params = [];
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var change = function(event) {
					var p = this.state.params.slice();
					p[idx] = event.target.value;
					this.setState({
						params: p,
						save: true,
					});
				}.bind(this);
				return (
					React.createElement("li", {key: idx}, 
						param, ": ", React.createElement("input", {type: "text", style: {width: '75%'}, onChange: change, value: this.state.params[idx]})
					)
				);
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(React.createElement("li", {key: 'oauth'}, React.createElement("a", {href: this.props.params.OAuthURL}, "connect")));
		}
		var save;
		if (this.state.save) {
			save = React.createElement("button", {onClick: this.save}, "save");
		}
		return (
			React.createElement("div", null, 
				React.createElement("ul", null, params), 
				save
			)
		);
	}
});

var ProtocolRow = React.createClass({displayName: "ProtocolRow",
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
		var icon = 'fa fa-fw fa-border fa-2x clickable ';
		return (
			React.createElement("tr", null, 
				React.createElement("td", null, this.props.protocol), 
				React.createElement("td", null, this.props.name), 
				React.createElement("td", null, React.createElement("i", {className: icon + 'fa-times', onClick: this.remove})), 
				React.createElement("td", null, React.createElement("i", {className: icon + 'fa-repeat', onClick: this.refresh}))
			)
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
			return 'add-' + t.ID.UID;
		});
		params.unshift('clear');
		POST('/api/playlist/change/' + name, mkcmd(params));
	},
	render: function() {
		return (
			React.createElement("div", null, 
				React.createElement("h4", null, "Queue"), 
				React.createElement("button", {onClick: this.clear}, "clear"), 
				" ", 
				React.createElement("button", {onClick: this.save}, "save"), 
				React.createElement(Tracks, {tracks: this.state.Queue, noIdx: true, isqueue: true})
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
		return (
			React.createElement("div", null, 
				React.createElement("h4", null, this.props.params.Playlist), 
				React.createElement("button", {onClick: this.clear}, "delete playlist"), 
				React.createElement(Tracks, {tracks: this.state.Playlists[this.props.params.Playlist], useIdxAsNum: true})
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
	mixins: [
		Reflux.listenTo(Stores.playlist, 'setState'),
		Reflux.listenTo(Stores.error, 'error')
	],
	componentDidMount: function() {
		this.startWS();
	},
	getInitialState: function() {
		return {};
	},
	startWS: function() {
		var ws = new WebSocket('ws://' + window.location.host + '/ws/');
		ws.onmessage = function(e) {
			this.setState({connected: true});
			var d = JSON.parse(e.data);
			if (Actions[d.Type]) {
				Actions[d.Type](d.Data);
			} else {
				console.log("missing action", d.Type);
			}
		}.bind(this);
		ws.onclose = function() {
			this.setState({connected: false});
			setTimeout(this.startWS, 1000);
		}.bind(this);
	},
	error: function(d) {
		this.setState({error: d});
	},
	clearError: function() {
		this.setState({error: null});
	},
	render: function() {
		var overlay;
		if (!this.state.connected) {
			overlay = (
				React.createElement("div", {id: "overlay"}, 
					React.createElement("div", {id: "overlay-text"}, 
						"mog lost connection with server", 
						React.createElement("p", null), 
						"attempting to reconnect..."
					)
				)
			);
		}
		var playlists = _.map(this.state.Playlists, function(_, key) {
			return React.createElement("li", {key: key}, React.createElement(Link, {to: "playlist", params: {Playlist: key}}, key));
		});
		var error;
		if (this.state.error) {
			var time = new Date(this.state.error.Time);
			error = React.createElement("div", null, React.createElement("a", {href: "#", onClick: this.clearError}, "[clear]"), " error at ", time.toString(), ": ", this.state.error.Error);
		}
		return (
			React.createElement("div", null, 
				overlay, 
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
					error, 
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
	mixins: [Reflux.listenTo(Stores.status, 'setStatus')],
	cmd: function(cmd) {
		return function() {
			POST('/api/cmd/' + cmd);
		};
	},
	getInitialState: function() {
		return {};
	},
	setStatus: function(d) {
		if (!this.state.Song || d.Song.UID != this.state.Song.UID) {
			Actions.active(d.Song.UID);
		}
		this.setState(d);
		var title = 'mog';
		if (this.state.SongInfo && this.state.SongInfo.Title) {
			title = this.state.SongInfo.Title + ' - ' + title;
		}
		if (this.state.State == 0) {
			title = '\u25B6 ' + title;
		}
		document.title = title;
	},
	seek: function(event) {
		if (!this.state.Time) {
			return;
		}
		var pos = event.screenX / window.innerWidth;
		var s = pos * this.state.Time;
		POST('/api/cmd/seek?pos=' + s + 'ns');
	},
	render: function() {
		var status;
		var pos = 0;
		if (this.state.Song && this.state.Song.ID) {
			var info = this.state.SongInfo;
			var song = this.state.Song.UID;
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
			pos = this.state.Elapsed / this.state.Time;
			pos = (pos * 100) + '%';
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
				React.createElement("div", {id: "seek", onClick: this.seek}, 
					React.createElement("div", {id: "seek-pos", style: {width: pos}})
				), 
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
