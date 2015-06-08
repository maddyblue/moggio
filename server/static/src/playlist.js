// @flow

var exports = module.exports = {};

var List = require('./list.js');
var Mog = require('./mog.js');
var React = require('react');
var Reflux = require('reflux');
var _ = require('underscore');
var mui = require('material-ui');

var { RaisedButton } = mui;

exports.Queue = React.createClass({
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	clear: function() {
		var params = Mog.mkcmd([
			'clear',
		]);
		Mog.POST('/api/queue/change', params);
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
		Mog.POST('/api/playlist/change/' + name, Mog.mkcmd(params));
	},
	render: function() {
		return (
			<div>
				<h2>Queue</h2>
				<RaisedButton onClick={this.clear} label='clear' />
				&nbsp;
				<RaisedButton onClick={this.save} label='save' />
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
		var params = Mog.mkcmd([
			'clear',
		]);
		Mog.POST('/api/playlist/change/' + this.props.params.Playlist, params);
	},
	render: function() {
		return (
			<div>
				<h2>{this.props.params.Playlist}</h2>
				<RaisedButton onClick={this.clear} label='delete playlist' />
				<List.Tracks tracks={this.state.Playlists[this.props.params.Playlist]} useIdxAsNum={true} />
			</div>
		);
	}
});
