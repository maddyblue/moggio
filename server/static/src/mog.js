/** @jsx React.DOM */

var Actions = Reflux.createActions([
	'protocols',
	'status',
	'tracks',
]);

var Stores = {};

_.each(Actions, function(action, name) {
	Stores[name] = Reflux.createStore({
		init: function() {
			this.listenTo(action, this.trigger);
		}
	});
});

// tracklist

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
	mixins: [Reflux.listenTo(Stores.tracks, 'setTracks')],
	getInitialState: function() {
		return {
			tracks: {},
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
			return <Track key={key} {...t} />;
		}));
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

// protocols

var Protocols = React.createClass({
	mixins: [Reflux.listenTo(Stores.protocols, 'setState')],
	getInitialState: function() {
		return {
			Available: {},
			Current: {},
			Selected: 'file',
		};
	},
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return <option key={protocol}>{protocol}</option>;
		}.bind(this));
		var protocols = [];
		_.each(this.state.Current, function(instances, protocol) {
			_.each(instances, function(inst, key) {
				protocols.push(<Protocol key={'current-' + protocol + '-' + key} protocol={protocol} params={this.state.Available[protocol]} instance={inst} name={key} />);
			}, this);
		}, this);
		var selected;
		if (this.state.Selected) {
			selected = <Protocol key={'selected-' + this.state.Selected} protocol={this.state.Selected} params={this.state.Available[this.state.Selected]} />;
		}
		return <div>
			<h2>New Protocol</h2>
			<select onChange={this.handleChange} value={this.state.Selected}>{options}</select>
			{selected}
			<h2>Existing Protocols</h2>
			{protocols}
		</div>;
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
	componentDidMount: function() {
		this.startWS();
	},
	startWS: function() {
		var ws = new WebSocket('ws://' + window.location.host + '/ws/');
		ws.onmessage = function(e) {
			var d = JSON.parse(e.data);
			if (d.Type != 'status') {
				console.log(d.Type);
			}
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

var Player = React.createClass({
	mixins: [Reflux.listenTo(Stores.status, 'setState')],
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
		if (!this.status) {
			status = <div>unknown</div>;
		} else {
			status = (
				<ul className="list-inline">
					<li>pl: {this.status.Playlist}</li>
					<li>state: {this.status.State}</li>
					<li>song: {this.status.Song}</li>
					<li>elapsed: <Time time={this.status.Elapsed} /></li>
					<li>time: <Time time={this.status.Time} /></li>
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