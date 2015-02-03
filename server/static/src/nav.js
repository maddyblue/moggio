// @flow

var Router = ReactRouter;
var Route = Router.Route;
var NotFoundRoute = Router.NotFoundRoute;
var DefaultRoute = Router.DefaultRoute;
var Link = Router.Link;
var RouteHandler = Router.RouteHandler;
var Redirect = Router.Redirect;

var App = React.createClass({
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
			return <li key={key}><Link to="playlist" params={{Playlist: key}}>{key}</Link></li>;
		});
		return (
			<div>
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
			status = (
				<span>
					<span>
						<Time time={this.state.Elapsed} /> /
						<Time time={this.state.Time} />
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
		<DefaultRoute handler={TrackList} />
		<Route name="album" path="/album/:Album" handler={Album} />
		<Route name="artist" path="/artist/:Artist" handler={Artist} />
		<Route name="playlist" path="/playlist/:Playlist" handler={Playlist} />
		<Route name="protocols" handler={Protocols} />
		<Route name="queue" handler={Queue} />
	</Route>
);

Router.run(routes, Router.HistoryLocation, function (Handler, state) {
	var params = state.params;
	React.render(<Handler params={params}/>, document.getElementById('main'));
});
