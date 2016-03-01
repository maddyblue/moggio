var exports = module.exports = {};

var List = require('./list.js');
var Moggio = require('./moggio.js');
var React = require('react');
var Reflux = require('reflux');
var _ = require('underscore');

var { Button } = require('./mdl.js');

exports.Queue = React.createClass({
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	clear: function() {
		var params = [['clear']];
		Moggio.POST('/api/queue/change', params);
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
			return ['add', t.ID.UID];
		});
		params.unshift(['clear']);
		Moggio.POST('/api/playlist/change/' + name, params);
	},
	render: function() {
		return (
			<div>
				<div className="mdl-typography--display-3 mdl-color-text--grey-600">Queue</div>
				<Button raised={true} onClick={this.clear}>clear</Button>
				&nbsp;
				<Button raised={true} onClick={this.save}>save</Button>
				<List.Tracks tracks={this.state.Queue} noIdx={true} isqueue={true} />
			</div>
		);
	}
});

exports.Playlist = React.createClass({
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
		var params = [['clear']];
		Moggio.POST('/api/playlist/change/' + this.props.params.Playlist, params);
	},
	render: function() {
		return (
			<div>
				<div className="mdl-typography--display-3 mdl-color-text--grey-600">{this.props.params.Playlist}</div>
				<Button raised={true} onClick={this.clear}>delete playlist</Button>
				<List.Tracks tracks={this.state.Playlists[this.props.params.Playlist]} useIdxAsNum={true} />
			</div>
		);
	}
});
