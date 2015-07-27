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

// todo: get these from the CSS somehow
var Colors = {
	grey300: '#e0e0e0',
	orange500: 'rgb(255, 152, 0)',
	grey500: 'rgb(158, 158, 158)',
};

var Mog = require('./mog.js');

var Group = require('./group.js');
var List = require('./list.js');
var Playlist = require('./playlist.js');
var Protocol = require('./protocol.js');

var { Button } = require('./mdl.js');

var App = React.createClass({
	mixins: [
		Reflux.listenTo(Stores.playlist, 'setState'),
		Reflux.listenTo(Stores.error, 'error'),
		Router.Navigation,
		Router.State,
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
	getSelectedIndex() {
		var currentItem;
		var menuItems = this.state.menuItems;
		for (var i = menuItems.length - 1; i >= 0; i--) {
			currentItem = menuItems[i];
			if (currentItem.route && this.context.router.isActive(currentItem.route)) {
				return i;
			}
		}
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
			return <Link key={k} className="mdl-navigation__link" to={v.route}>{v.text}</Link>;
		});
		var playlists;
		if (this.state.Playlists) {
			var entries = _.map(this.state.Playlists, function(_, key) {
				return <Link key={"playlist-" + key} className="mdl-navigation__link" to="playlist" params={{Playlist: key}}>{key}</Link>;
			});
			playlists = (
				<nav className="mdl-navigation">
					{entries}
				</nav>
			);
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
	openQueue: function() {
		this.transitionTo('queue');
	},
	render: function() {
		var title, album;
		var animation = '';
		var pos = 0;
		var dur = '0s';
		var img;
		if (this.state.Song && this.state.Song.ID) {
			var info = this.state.SongInfo;
			var song = this.state.Song.UID;
			title = <div style={{fontWeight: '500'}}>{info.Title}</div>;
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
			pos = this.state.Elapsed / this.state.Time * 100;
			dur = (this.state.Time - this.state.Elapsed) / 1e9 + 's';
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
				img = <div style={_.extend({backgroundColor: Colors.grey300}, istyle)} />;
			}
		};
		var play = this.state.State == 0 ? 'pause' : 'play_circle_filled';
		// TODO: dynamically set this to the correct CSS color/class
		var primary = 'red';
		var repeat = this.state.Repeat ? primary : {};
		var random = this.state.Random ? primary : {};
		var ctrlStyle = {
			position: 'absolute',
			left: '50%',
			width: '240px',
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
		var ad = animation + ' ' + dur;
		var seekPosStyle = {
			position: 'absolute',
			bottom: '0',
			top: '0',
			left: '0',
			animationTimingFunction: 'linear',
			width: animation == '' ? 0 : '100%',
			animation: ad,
			backgroundColor: Colors.orange500,
		};
		return (
			<div>
				<div id="seek" onClick={this.seek}>
					<div style={{position: 'absolute', left: '0', width: pos + '%', bottom: '0', top: '0', backgroundColor: seekPosStyle.backgroundColor}}/>
					<div style={{position: 'absolute', right: '0', width: (100 - pos) + '%', bottom: '0', top: '0', backgroundColor: Colors.grey500}}>
						<div style={seekPosStyle}/>
					</div>
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
					<div style={ctrlStyle}>
						<Button onClick={this.cmd('repeat')} style={_.extend({}, btnStyle, repeat)} icon={true}>
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
						<Button onClick={this.cmd('random')} style={_.extend({}, btnStyle, random)} icon={true}>
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
