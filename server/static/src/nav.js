// @flow

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
			POST('/api/cmd/' + cmd);
		};
	},
	getInitialState: function() {
		return {};
	},
	render: function() {
		var status;
		if (!this.state.Song) {
			status = <span>unknown</span>;
		} else {
			status = (
				<span>
					<span>pl: {this.state.Playlist}</span>
					<span>state: {this.state.State}</span>
					<span>elapsed: <Time time={this.state.Elapsed} /></span>
					<span>time: <Time time={this.state.Time} /></span>
					<span>song: {this.state.Song}</span>
				</span>
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
			<div>
				<span><button onClick={this.cmd('prev')}>⇤</button></span>
				<span><button onClick={this.cmd('pause')}>{play}</button></span>
				<span><button onClick={this.cmd('next')}>⇥</button></span>
				<span>{status}</span>
			</div>
		);
	}
});

var player = <Player />;
React.render(player, document.getElementById('player'));