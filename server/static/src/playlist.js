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
			return 'add-' + t.UID;
		});
		params.unshift('clear');
		POST('/api/playlist/change/' + name, mkcmd(params));
	},
	render: function() {
		var q = _.map(this.state.Queue, function(val) {
			return {
				ID: val
			};
		});
		return (
			<div>
				<h4>Queue</h4>
				<button onClick={this.clear}>clear</button>
				<button onClick={this.save}>save</button>
				<Tracks tracks={q} isqueue={true} />
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
		var params = mkcmd([
			'clear',
		]);
		POST('/api/playlist/change/' + this.props.params.Playlist, params);
	},
	render: function() {
		var q = _.map(this.state.Playlists[this.props.params.Playlist], function(val) {
			return {
				ID: val
			};
		});
		return (
			<div>
				<h4>{this.props.params.Playlist}</h4>
				<Tracks tracks={q} useIdxAsNum={true} />
			</div>
		);
	}
});
