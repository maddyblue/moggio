/** @jsx React.DOM */

var TrackListRow = React.createClass({
	render: function() {
		return (<tr><td>{this.props.protocol}</td><td>{this.props.id}</td></tr>);
	}
});

var Track = React.createClass({
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
			<tr>
				<td><button onClick={this.play}>&#x25b6;</button></td>
				<td>{this.props.key}</td>
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
			this.setState({tracks: result});
		}.bind(this));
	},
	render: function() {
		var tracks = this.state.tracks.map(function (t) {
			return <Track key={t} />;
		});
		return (
			<table>
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
				{this.props.key} <input type="text" onChange={this.paramChange} value={this.state.value} />
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
		var params = this.props.params.map(function(param, idx) {
			var current = this.props.current || [];
			return <ProtocolParam key={param} ref={idx} value={current[idx]} change={this.setSave} />;
		}.bind(this));
		var save;
		if (this.state.save) {
			save = <button onClick={this.save}>save</button>;
		}
		return (
			<div key={this.props.key}>
				<h2>{this.props.key}</h2>
				<ul>{params}</ul>
				{save}
			</div>
		);
	}
});

var Protocols = React.createClass({
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
			return <Protocol key={protocol} params={this.state.available[protocol]} current={this.state.current[protocol]} />;
		}.bind(this));
		return <div>{protocols}</div>;
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
			<ul>
				<Link href="/list" name="List" handler={TrackList} index={true} />
				<Link href="/protocols" name="Protocols" handler={Protocols} />
			</ul>
		);
	}
});

React.renderComponent(<Navigation />, document.getElementById('navigation'));

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

var Player = React.createClass({
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
			<div>
				<button onClick={this.cmd('prev')}>prev</button>
				<button onClick={this.cmd('pause')}>play/pause</button>
				<button onClick={this.cmd('next')}>next</button>
			</div>
		);
		var status;
		if (!this.state.status) {
			status = <div>unknown</div>;
		} else {
			status = (
				<ul>
					<li>cache: {this.state.cache}</li>
					<li>pl: {this.state.status.Playlist}</li>
					<li>state: {this.state.status.State}</li>
					<li>song: {this.state.status.Song}</li>
					<li>elapsed: {this.state.status.Elapsed}</li>
					<li>time: {this.state.status.Time}</li>
				</ul>
			);
		};
		return <div>{player}{status}</div>;
	}
});

React.renderComponent(<Player />, document.getElementById('player'));