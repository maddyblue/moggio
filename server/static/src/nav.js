// @flow

var React = require('react');
var Reflux = require('reflux');
var Router = require('react-router');
var _ = require('underscore');

var Route = Router.Route;
var NotFoundRoute = Router.NotFoundRoute;
var DefaultRoute = Router.DefaultRoute;
var Link = Router.Link;
var RouteHandler = Router.RouteHandler;
var Redirect = Router.Redirect;

var Mog = require('./mog.js');
var List = require('./list.js');
var Playlist = require('./playlist.js');
var Protocol = require('./protocol.js');

var App = React.createClass({
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
				<div id="overlay">
					<div id="overlay-text">
						mog lost connection with server
						<p/>
						attempting to reconnect...
					</div>
				</div>
			);
		}
		var playlists = _.map(this.state.Playlists, function(_, key) {
			return <li key={key}><Link to="playlist" params={{Playlist: key}}>{key}</Link></li>;
		});
		var error;
		if (this.state.error) {
			var time = new Date(this.state.error.Time);
			error = <div><a href="#" onClick={this.clearError}>[clear]</a> error at {time.toString()}: {this.state.error.Error}</div>;
		}
		return (
			<div>
				{overlay}
				<header>
					<ul>
						<li><Link to="app">Music</Link></li>
						<li><Link to="protocols">Sources</Link></li>
						<li><Link to="queue">Queue</Link></li>
					</ul>
					<h4>Playlists</h4>
					<ul>{playlists}</ul>
				</header>
				<main>
					{error}
					<RouteHandler {...this.props}/>
				</main>
				<footer>
					<Player/>
				</footer>
			</div>
		);
	}
});

var Player = React.createClass({
	mixins: [Reflux.listenTo(Stores.status, 'setStatus')],
	cmd: function(cmd) {
		return function() {
			Mog.POST('/api/cmd/' + cmd);
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
		Mog.POST('/api/cmd/seek?pos=' + s + 'ns');
	},
	render: function() {
		var status;
		var pos = 0;
		if (this.state.Song && this.state.Song.ID) {
			var info = this.state.SongInfo;
			var song = this.state.Song.UID;
			var album;
			if (info.Album) {
				album = <span>- <Link to="album" params={info}>{info.Album}</Link></span>;
			}
			var artist;
			if (info.Artist) {
				artist = <span>- <Link to="artist" params={info}>{info.Artist}</Link></span>;
			}
			song = (
				<span>
					{info.Title}
					{album}
					{artist}
				</span>
			);
			pos = this.state.Elapsed / this.state.Time;
			pos = (pos * 100) + '%';
			status = (
				<span>
					<span>
						<Mog.Time time={this.state.Elapsed} /> /
						<Mog.Time time={this.state.Time} />
					</span>
					{song}
				</span>
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
			<div>
				<div id="seek" onClick={this.seek}>
					<div id="seek-pos" style={{width: pos}}/>
				</div>
				<span><i className={icon + repeat + 'fa-repeat'} onClick={this.cmd('repeat')} /></span>
				<span><i className={icon + 'fa-fast-backward'} onClick={this.cmd('prev')} /></span>
				<span><i className={icon + play} onClick={this.cmd('pause')} /></span>
				<span><i className={icon + 'fa-fast-forward'} onClick={this.cmd('next')} /></span>
				<span><i className={icon + random + 'fa-random'} onClick={this.cmd('random')} /></span>
				<span>{status}</span>
			</div>
		);
	}
});

var routes = (
	<Route name="app" path="/" handler={App}>
		<DefaultRoute handler={List.TrackList} />
		<Route name="album" path="/album/:Album" handler={List.Album} />
		<Route name="artist" path="/artist/:Artist" handler={List.Artist} />
		<Route name="playlist" path="/playlist/:Playlist" handler={Playlist.Playlist} />
		<Route name="protocols" handler={Protocol.Protocols} />
		<Route name="queue" handler={Playlist.Queue} />
	</Route>
);

Router.run(routes, Router.HistoryLocation, function (Handler, state) {
	var params = state.params;
	React.render(<Handler params={params}/>, document.getElementById('main'));
});
