// @flow

var React = require('react');
var Reflux = require('reflux');
var Router = require('react-router');
var _ = require('underscore');
var mui = require('material-ui');
var ThemeManager = new mui.Styles.ThemeManager();

var injectTapEventPlugin = require('react-tap-event-plugin');
injectTapEventPlugin();

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

var { AppBar, LeftNav, MenuItem, IconButton, FloatingActionButton } = mui;

var App = React.createClass({
	mixins: [
		Reflux.listenTo(Stores.playlist, 'setState'),
		Reflux.listenTo(Stores.error, 'error'),
		Router.Navigation
	],
	childContextTypes: {
		muiTheme: React.PropTypes.object
	},
	getChildContext: function() {
		return {
			muiTheme: ThemeManager.getCurrentTheme()
		};
	},
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
	leftNavToggle: function() {
		this.refs.leftNav.toggle();
	},
	leftNavChange: function(e, selectedIndex, menuItem) {
		// TODO: correctly set navIndex on initial load
		this.setState({navIndex: selectedIndex});
		this.transitionTo(menuItem.route, menuItem.params);
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
		var menuItems = _.clone(navMenuItems);
		_.each(this.state.Playlists, function(_, key) {
			menuItems.push({
				route: 'playlist',
				params: {Playlist: key},
				text: key
			});
		});
		var error;
		if (this.state.error) {
			var time = new Date(this.state.error.Time);
			error = <div><a href="#" onClick={this.clearError}>[clear]</a> error at {time.toString()}: {this.state.error.Error}</div>;
		}
		return (
			<div>
				{overlay}
				{error}
				<AppBar
					title={'mog'}
					onLeftIconButtonTouchTap={this.leftNavToggle}
					/>
				<LeftNav
					ref="leftNav"
					docked={false}
					menuItems={menuItems}
					onChange={this.leftNavChange}
					selectedIndex={this.state.navIndex}
					/>
				<RouteHandler {...this.props}/>
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
	{ type: MenuItem.Types.SUBHEADER, text: 'Playlists' },
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
		var pos = 0;
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
			pos = this.state.Elapsed / this.state.Time;
			pos = (pos * 100) + '%';
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
				img = <div style={_.extend({backgroundColor: 'white'}, istyle)} />;
			}
		};
		var play = this.state.State == 0 ? 'pause' : 'play_circle_filled';
		var primary = {color: ThemeManager.getCurrentTheme().palette.accent1Color};
		var repeat = this.state.Repeat ? primary : {};
		var random = this.state.Random ? primary : {};
		var ctrlStyle = {
			position: 'absolute',
			left: '50%',
			width: '240px',
			transform: 'translateX(-50%)',
			bottom: '0',
			height: '70px',
			textAlign: 'center'
		};
		var btnStyle = {
			position: 'relative',
			top: '50%',
			transform: 'translateY(-50%)',
			backgroundColor: '#f5f5f5'
		};
		var statusStyle = {
			position: 'absolute',
			left: '5px',
			textAlign: 'left',
			top: '20px',
			width: '50%',
			whiteSpace: 'nowrap',
			overflow: 'hidden'
		};
		var rightStyle = {
			position: 'absolute',
			right: '5px',
			height: '70px'
		};
		return (
			<div>
				<div id="seek" onClick={this.seek}>
					<div id="seek-pos" style={{width: pos}}/>
				</div>
				{img}
				<div style={{position: 'absolute', left: '80px', bottom: '0', right: '0', height: '70px', textAlign: 'center'}}>
					<div style={statusStyle}>
						{title}
						{album}
					</div>
					<div style={rightStyle}>
						<IconButton onClick={this.openQueue} style={btnStyle}>
							<i className='material-icons'>queue_music</i>
						</IconButton>
					</div>
					<div style={ctrlStyle}>
						<IconButton onClick={this.cmd('repeat')} style={_.extend({}, btnStyle, repeat)}>
							<i className='material-icons'>repeat</i>
						</IconButton>
						<IconButton onClick={this.cmd('prev')} style={btnStyle}>
							<i className='material-icons'>skip_previous</i>
						</IconButton>
						<IconButton onClick={this.cmd('pause')} style={btnStyle}>
							<i className='material-icons'>{play}</i>
						</IconButton>
						<IconButton onClick={this.cmd('next')} style={btnStyle}>
							<i className='material-icons'>skip_next</i>
						</IconButton>
						<IconButton onClick={this.cmd('random')} style={_.extend({}, btnStyle, random)}>
							<i className='material-icons'>shuffle</i>
						</IconButton>
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
