// @flow

var Queue = React.createClass({
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	clear: function() {
		var params = mkcmd([
			'clear',
		]);
		POST('/api/queue/change', params);
	},
	save: function() {
		var name = prompt("Playlist name:");
		if (!name) {
			return;
		}
		if (this.state.Playlists[name]) {
			if (!window.confirm("Overwrite existing playlist?")) {
				return;
			}
		}
		var params = _.map(this.state.Queue, function(t) {
			return 'add-' + t.ID.UID;
		});
		params.unshift('clear');
		POST('/api/playlist/change/' + name, mkcmd(params));
	},
	render: function() {
		return (
			<div>
				<h4>Queue</h4>
				<button onClick={this.clear}>clear</button>
				<button onClick={this.save}>save</button>
				<Tracks tracks={this.state.Queue} noIdx={true} isqueue={true} />
			</div>
		);
	}
});

var Playlist = React.createClass({
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {
			Playlists: {}
		};
	},
	clear: function() {
		if (!confirm("Delete playlist?")) {
			return;
		}
		var params = mkcmd([
			'clear',
		]);
		POST('/api/playlist/change/' + this.props.params.Playlist, params);
	},
	render: function() {
		return (
			<div>
				<h4>{this.props.params.Playlist}</h4>
				<button onClick={this.clear}>delete playlist</button>
				<Tracks tracks={this.state.Playlists[this.props.params.Playlist]} useIdxAsNum={true} />
			</div>
		);
	}
});
