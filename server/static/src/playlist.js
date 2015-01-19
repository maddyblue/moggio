// @flow

var Playlist = React.createClass({
	mixins: [Reflux.listenTo(Stores.playlist, 'setState')],
	getInitialState: function() {
		return Stores.playlist.data || {};
	},
	render: function() {
		var tracks = _.map(this.state.Queue, (function (id) {
			return <Track key={id.UID} id={id}/>;
		}));
		return (
			<table className="table">
				<thead>
					<tr>
						<th>Name</th>
						<th>Time</th>
						<th>Artist</th>
						<th>Album</th>
					</tr>
				</thead>
				<tbody>{tracks}</tbody>
			</table>
		);
	}
});
