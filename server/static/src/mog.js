/** @jsx React.DOM */

var TrackListRow = React.createClass({
	render: function() {
		return (<tr><td>{this.props.protocol}</td><td>{this.props.id}</td></tr>);
	}
});

var Time = React.createClass({
	render: function() {
		var t = moment.duration(this.props.time / 1e6);
		var s = t.seconds().toString();
		if (s.length == 1) {
			s = "0" + s;
		}
		return <span>{t.minutes()}:{s}</span>;
	}
});

var Track = React.createClass({
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
			<tr>
				<td><button className="btn btn-default btn-sm" onClick={this.play}>&#x25b6;</button> {this.props.Info.Title}</td>
				<td><Time time={this.props.Info.Time} /></td>
				<td>{this.props.Info.Artist}</td>
				<td>{this.props.Info.Album}</td>
			</tr>
		);
	}
});

var TrackList = React.createClass({
	getInitialState: function() {
		return {
			tracks: []
		};
	},
	componentDidMount: function() {
		$.get('/api/list', function(result) {
			result.forEach(function(t) {
				t.ID.uid = t.ID.Protocol + "|" + t.ID.Key + "|" + t.ID.ID;
			});
			this.setState({tracks: result});
		}.bind(this));
	},
	render: function() {
		var tracks = this.state.tracks.map(function (t) {
			return <Track key={t.ID.uid} {...t} />;
		});
		return (
			<table className="table">
				<thead>
					<tr>
						<th>Name</th>
						<th>Time</th>
						<th>Artist</th>
						<th>Album</th>
					</tr>
				</thead>
				<tbody>{tracks}</tbody>
			</table>
		);
	}
});

var ProtocolParam = React.createClass({
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
			<li>
				{this.props.key} <input type="text" onChange={this.paramChange} value={this.state.value || this.props.value} disabled={this.props.disabled ? 'disabled' : ''} />
			</li>
		);
	}
});

var ProtocolOAuth = React.createClass({
	render: function() {
		var token;
		if (this.props.token) {
			token = <div>Connected until {this.props.token.expiry}</div>;
		}
		return (
			<li>
				{token}
				<a href={this.props.url}>connect</a>
			</li>
		);
	}
});

var Protocol = React.createClass({
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
				return <ProtocolParam key={param} ref={idx} value={current[idx]} change={this.setSave} disabled={disabled} />;
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(<ProtocolOAuth key={'oauth-' + this.props.key} url={this.props.params.OAuthURL} token={this.props.instance.OAuthToken} disabled={disabled} />);
		}
		var save;
		if (this.state.save) {
			save = <button onClick={this.save}>save</button>;
		}
		var title;
		if (this.props.name) {
			title = <h3>{this.props.protocol}: {this.props.name}
					<small><button onClick={this.remove}>remove</button></small>
				</h3>;
		}
		return <div>
				{title}
				<ul>{params}</ul>
				{save}
			</div>;
	}
});

var Protocols = React.createClass({
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
			return <option key={protocol}>{protocol}</option>;
		}.bind(this));
		var protocols = [];
		_.each(this.state.current, function(instances, protocol) {
			_.each(instances, function(inst, key) {
				protocols.push(<Protocol key={'current-' + protocol + '-' + key} protocol={protocol} params={this.state.available[protocol]} instance={inst} name={key} />);
			}, this);
		}, this);
		var selected;
		if (this.state.selected) {
			selected = <Protocol key={'selected-' + this.state.selected} protocol={this.state.selected} params={this.state.available[this.state.selected]} />;
		}
		return <div>
			<h2>New Protocol</h2>
			<select onChange={this.handleChange} value={this.state.selected}>{options}</select>
			{selected}
			<h2>Existing Protocols</h2>
			{protocols}
		</div>;
	}
});

var routes = {};

var Link = React.createClass({
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
		return <li><a href={this.props.href} onClick={this.click}>{this.props.name}</a></li>
	}
});

var Navigation = React.createClass({
	render: function() {
		return (
			<ul className="nav navbar-nav">
				<Link href="/list" name="List" handler={TrackList} index={true} />
				<Link href="/protocols" name="Protocols" handler={Protocols} />
			</ul>
		);
	}
});

var navigation = <Navigation />;
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

var Player = React.createClass({
	getInitialState: function() {
		return {};
	},
	componentDidUpdate: function(props, state) {
		if (state.status && state.status.Song && state.status.Song.ID) {
			getSong(this.state.cache, state.status.Song, this.setState.bind(this));
		}
	},
	startWS: function() {
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
		var status;
		if (!this.state.status) {
			status = <div>unknown</div>;
		} else {
			status = (
				<ul className="list-inline">
					<li>cache: {this.state.cache}</li>
					<li>pl: {this.state.status.Playlist}</li>
					<li>state: {this.state.status.State}</li>
					<li>song: {this.state.status.Song}</li>
					<li>elapsed: <Time time={this.state.status.Elapsed} /></li>
					<li>time: <Time time={this.state.status.Time} /></li>
				</ul>
			);
		};
		return (
			<ul className="list-inline">
				<li><button className="btn btn-default" onClick={this.cmd('prev')}>prev</button></li>
				<li><button className="btn btn-default" onClick={this.cmd('pause')}>play/pause</button></li>
				<li><button className="btn btn-default" onClick={this.cmd('next')}>next</button></li>
				<li>{status}</li>
			</ul>
		);
	}
});

var player = <Player />;
React.render(player, document.getElementById('player'));
