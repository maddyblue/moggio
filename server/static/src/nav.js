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

var Group = require('./group.js');
var List = require('./list.js');
var Playlist = require('./playlist.js');
var Protocol = require('./protocol.js');

var { Button } = require('./mdl.js');

var App = React.createClass({
	mixins: [
		Reflux.listenTo(Stores.error, 'error'),
		Reflux.listenTo(Stores.playlist, 'setState'),
		Reflux.listenTo(Stores.status, 'setState'),
		Router.Navigation,
		Router.State,
	],
	componentDidMount: function() {
		this.startWS();
		var that = this;
		fetch('https://api.github.com/repos/mjibson/mog/releases/latest')
		.then(function (r) {
			r.json().then(function(j) {
				var v = j.tag_name.substr(1);
				if (j.tag_name == mogVersion) {
					return;
				}
				that.setState({update: {
					name: v,
					link: j.html_url
				}});
			});
		});
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
	routeClass: function(route, params) {
		var active = this.context.router.isActive(route, params);
		var path = this.context.router.getCurrentPath();
		if (route == 'app' && path != '/') {
			active = false;
		}
		return active ? ' mdl-color-text--accent' : '';
	},
	logout: function(event) {
		event.preventDefault();
		Mog.POST('/api/token/register');
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
		var error;
		if (this.state.error) {
			var time = new Date(this.state.error.Time);
			error = (
				<div style={{padding: '10px'}}>
					<Button onClick={this.clearError} raised={true} primary={true}>clear</Button>
					<span style={{paddingLeft: '10px'}}>
						error at {time.toString()}: {this.state.error.Error}
					</span>
				</div>
			);
		}
		var menuItems = _.map(navMenuItems, function(v, k) {
			return <Link key={k} className={"mdl-navigation__link" + this.routeClass(v.route)} to={v.route}>{v.text}</Link>;
		}.bind(this));
		if (this.state.CentralURL) {
			if (this.state.Username) {
				var un = (
					<span key="username" className="mdl-navigation__link">
						{this.state.Username}
						<br/>
						<a href="" onClick={this.logout}>[logout]</a>
					</span>
				);
				menuItems.unshift(un);
			} else {
				var origin = location.protocol + '//' + location.host + '/api/token/register';
				var params = "?redirect=" + encodeURIComponent(origin);
				if (this.state.Hostname) {
					params += '&hostname=' + encodeURIComponent(this.state.Hostname);
				}
				menuItems.unshift(<a key="username" href={this.state.CentralURL + '/token' + params} className="mdl-navigation__link">login</a>);
			}
		}
		var playlists;
		if (this.state.Playlists) {
			var entries = _.map(this.state.Playlists, function(_, key) {
				var params = {Playlist: key};
				return <Link key={"playlist-" + key} className={"mdl-navigation__link" + this.routeClass('playlist', params)} to="playlist" params={params}>{key}</Link>;
			}.bind(this));
			playlists = (
				<nav className="mdl-navigation">
					{entries}
				</nav>
			);
		}
		var update;
		if (this.state.update) {
			update = <span className="mdl-layout-title">
				new version:&nbsp;
				<a href={this.state.update.link}>{this.state.update.name}</a>
			</span>;
		}
		return (
			<div>
				{overlay}
				<div className="top-main">
				<div className="mdl-layout mdl-js-layout mdl-layout--fixed-drawer
					mdl-layout--overlay-drawer-button">
					<div className="mdl-layout__drawer">
						<span className="mdl-layout-title">mog</span>
						<nav className="mdl-navigation">
							{menuItems}
						</nav>
						<span className="mdl-layout-title">Playlists</span>
						{playlists}
						{update}
					</div>
					<main className="mdl-layout__content">
						<div className="page-content">
							{error}
							<RouteHandler {...this.props} />
						</div>
					</main>
				</div>
				</div>

				<footer>
					<Player/>
				</footer>
			</div>
		);
	}
});

var navMenuItems = [
	{ route: 'app', text: 'Music' },
	{ route: 'protocols', text: 'Sources' },
	{ route: 'queue', text: 'Queue' },
	{ route: 'artists', text: 'Artists' },
	{ route: 'albums', text: 'Albums' },
];

var Player = React.createClass({
	mixins: [
		Reflux.listenTo(Stores.status, 'setStatus'),
		Router.Navigation
	],
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
		d.songStart = new Date() - d.Elapsed / 1e6;
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
		var offset = 80;
		var pos = (event.clientX - offset) / (window.innerWidth - offset);
		var s = pos * this.state.Time;
		Mog.POST('/api/cmd/seek?pos=' + s + 'ns');
	},
	openQueue: function() {
		this.transitionTo('queue');
	},
	renderSeek: function() {
		if (this.state.State == 0) {
			window.clearTimeout(this.timeout);
			this.timeout = setTimeout(function() {
				window.requestAnimationFrame(this.renderSeek);
			}.bind(this), 200);
		}
		var s = document.getElementById('seek-pos');
		var pos = 0;
		if (this.state.songStart) {
			var d = new Date - this.state.songStart;
			pos = d / this.state.Time * 1e8;
		}
		s.style.width = pos + '%';
	},
	render: function() {
		var title, album;
		var img;
		if (this.state.Song && this.state.Song.ID) {
			var info = this.state.SongInfo;
			var song = this.state.Song.UID;
			title = <div style={{fontWeight: '500'}}>{info.SongTitle || info.Title}</div>;
			var ialbum, iartist, joiner;
			if (info.Album) {
				ialbum = <Link to="album" params={info}>{info.Album}</Link>;
				if (info.Artist) {
					joiner = ' - ';
				}
			}
			if (info.Artist) {
				iartist = <Link to="artist" params={info}>{info.Artist}</Link>;
			}
			album = (
				<div>
					{iartist}
					{joiner}
					{ialbum}
				</div>
			);
			if (this.state.State == 0) {
				// If a song is playing, swap the animation (they are identical). This
				// triggers an animation restart which we need on song change.
				this.toggle = !this.toggle;
				var anicount = this.toggle ? '1' : '2';
				animation = 'seek' + anicount;
			}
			var istyle = {
				height: '80px',
				width: '80px',
				position: 'absolute',
				bottom: '0',
				left: '0',
			};
			if (info.ImageURL) {
				img = <img style={istyle} src={info.ImageURL}/>;
			} else {
				img = <div style={istyle} className="mdl-color--grey-300"/>;
			}
		};
		var play = this.state.State == 0 ? 'pause' : 'play_circle_filled';
		var ctrlStyle = {
			position: 'absolute',
			left: '50%',
			transform: 'translateX(-50%)',
			bottom: '0',
			height: '70px',
			textAlign: 'center',
		};
		var btnStyle = {
			position: 'relative',
			top: '50%',
			transform: 'translateY(-50%)',
		};
		var statusStyle = {
			position: 'absolute',
			left: '5px',
			textAlign: 'left',
			top: '16px',
			width: '50%',
			whiteSpace: 'nowrap',
			overflow: 'hidden',
		};
		var rightStyle = {
			position: 'absolute',
			right: '5px',
			height: '70px',
		};
		window.requestAnimationFrame(this.renderSeek);
		return (
			<div>
				<div id="seek" className="mdl-color--grey-500" onClick={this.seek}>
					<div id="seek-pos" className="mdl-color--orange-500" />
				</div>
				{img}
				<div style={{position: 'absolute', left: '80px', bottom: '0', right: '0', height: '70px', textAlign: 'center'}}>
					<div style={statusStyle}>
						{title}
						{album}
					</div>
					<div style={rightStyle}>
						<Button onClick={this.openQueue} style={btnStyle} icon={true}>
							<i className='material-icons'>queue_music</i>
						</Button>
					</div>
					<div style={ctrlStyle} className="mdl-color--grey-100">
						<Button onClick={this.cmd('repeat')} style={btnStyle} accent={this.state.Repeat} icon={true}>
							<i className='material-icons'>repeat</i>
						</Button>
						<Button onClick={this.cmd('prev')} style={btnStyle} icon={true}>
							<i className='material-icons'>skip_previous</i>
						</Button>
						<Button onClick={this.cmd('pause')} style={btnStyle} accent={true} icon={true}>
							<i className='material-icons'>{play}</i>
						</Button>
						<Button onClick={this.cmd('next')} style={btnStyle} icon={true}>
							<i className='material-icons'>skip_next</i>
						</Button>
						<Button onClick={this.cmd('random')} style={btnStyle} accent={this.state.Random} icon={true}>
							<i className='material-icons'>shuffle</i>
						</Button>
					</div>
				</div>
			</div>
		);
	}
});

var routes = (
	<Route name="app" path="/" handler={App}>
		<DefaultRoute handler={List.TrackList} />
		<Route name="album" path="/album/:Album" handler={List.Album} />
		<Route name="albums" path="/albums" handler={Group.Albums} />
		<Route name="artist" path="/artist/:Artist" handler={List.Artist} />
		<Route name="artists" path="/artists" handler={Group.Artists} />
		<Route name="playlist" path="/playlist/:Playlist" handler={Playlist.Playlist} />
		<Route name="protocols" handler={Protocol.Protocols} />
		<Route name="queue" handler={Playlist.Queue} />
	</Route>
);

Router.run(routes, Router.HistoryLocation, function (Handler, state) {
	var params = state.params;
	React.render(<Handler params={params}/>, document.getElementById('main'));
});
