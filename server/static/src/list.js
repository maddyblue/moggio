var TrackListRow = React.createClass({
	render: function() {
		return (<tr><td>{this.props.protocol}</td><td>{this.props.id}</td></tr>);
	}
});

var Time = React.createClass({
	render: function() {
		var t = moment.duration(this.props.time / 1e6);
		var s = t.seconds().toString();
		if (s.length == 1) {
			s = "0" + s;
		}
		return <span>{t.minutes()}:{s}</span>;
	}
});

var Track = React.createClass({
	play: function() {
		var params = {
			"clear": true,
			"add": JSON.stringify(this.props.ID)
		};
		$.get('/api/playlist/change?' + $.param(params))
			.success(function() {
				$.get('/api/cmd/play');
			});
	},
	render: function() {
		return (
			<tr>
				<td><button className="btn btn-default btn-sm" onClick={this.play}>&#x25b6;</button> {this.props.Info.Title}</td>
				<td><Time time={this.props.Info.Time} /></td>
				<td>{this.props.Info.Artist}</td>
				<td>{this.props.Info.Album}</td>
			</tr>
		);
	}
});

var TrackList = React.createClass({
	mixins: [Reflux.listenTo(Stores.tracks, 'setTracks')],
	getInitialState: function() {
		return {
			tracks: {},
		};
	},
	setTracks: function(tracks) {
		var sc = {};
		tracks.forEach(function(t) {
			var uid = t.ID.Protocol + "|" + t.ID.Key + "|" + t.ID.ID;
			sc[uid] = t;
		});
		this.setState({tracks: sc});
	},
	render: function() {
		var tracks = _.map(this.state.tracks, (function (t, key) {
			return <Track key={key} {...t} />;
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
