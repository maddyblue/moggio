// @flow

var Playlist = React.createClass({
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
	render: function() {
		var q = _.map(this.state.Queue, function(val) {
			return {
				ID: val
			};
		});
		return (
			<div>
				<button onClick={this.clear}>clear</button>
				<Tracks tracks={q} />
			</div>
		);
	}
});
