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
	getInitialState: function() {
		return {};
	},
	render: function() {
		var status;
		if (!this.state.status) {
			status = <span>unknown</span>;
		} else {
			status = (
				<span>
					<span>pl: {this.state.status.Playlist}</span>
					<span>state: {this.state.status.State}</span>
					<span>song: {this.state.status.Song}</span>
					<span>elapsed: <Time time={this.state.status.Elapsed} /></span>
					<span>time: <Time time={this.state.status.Time} /></span>
				</span>
			);
		};
		return (
			<div>
				<span><button onClick={this.cmd('prev')}>prev</button></span>
				<span><button onClick={this.cmd('pause')}>play/pause</button></span>
				<span><button onClick={this.cmd('next')}>next</button></span>
				<span>{status}</span>
			</div>
		);
	}
});

var player = <Player />;
React.render(player, document.getElementById('player'));