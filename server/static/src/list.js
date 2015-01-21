// @flow

var Track = React.createClass({
	mixins: [Reflux.listenTo(Stores.tracks, 'update')],
	play: function() {
		var params = mkcmd([
			'clear',
			'add-' + this.props.id.UID
		]);
		POST('/api/queue/change', params, function() {
			POST('/api/cmd/play');
		});
	},
	getInitialState: function() {
		if (this.props.info) {
			return {
				info: this.props.info
			};
		}
		var d = Lookup(this.props.id);
		if (d) {
			return {
				info: d.Info
			};
		}
		return {};
	},
	update: function() {
		this.setState(this.getInitialState());
	},
	render: function() {
		var info = this.state.info;
		if (!info) {
			return (
				<tr>
					<td>{this.props.id}</td>
				</tr>
			);
		}
		return (
			<tr>
				<td><button className="btn btn-default btn-sm" onClick={this.play}>&#x25b6;</button> {info.Title}</td>
				<td><Time time={info.Time} /></td>
				<td>{info.Artist}</td>
				<td>{info.Album}</td>
			</tr>
		);
	}
});

var Tracks = React.createClass({
	render: function() {
		var tracks = _.map(this.props.tracks, (function (t) {
			return <Track key={t.ID.UID} id={t.ID} info={t.Info} />;
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

var TrackList = React.createClass({
	mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
	getInitialState: function() {
		return Stores.tracks.data || {};
	},
	render: function() {
		return <Tracks tracks={this.state.Tracks} />;
	}
});
