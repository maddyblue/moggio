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
				<td><Link to="artist" params={info}>{info.Artist}</Link></td>
				<td><Link to="album" params={info}>{info.Album}</Link></td>
			</tr>
		);
	}
});

var Tracks = React.createClass({
	mkparams: function() {
		return _.map(this.props.tracks, function(t) {
			return 'add-' + t.ID.UID;
		});
	},
	play: function() {
		var params = this.mkparams();
		params.unshift('clear');
		POST('/api/queue/change', mkcmd(params), function() {
			POST('/api/cmd/play');
		});
	},
	add: function() {
		var params = this.mkparams();
		POST('/api/queue/change', mkcmd(params));
	},
	render: function() {
		var tracks = _.map(this.props.tracks, function(t, idx) {
			return <Track key={idx + '-' + t.ID.UID} id={t.ID} info={t.Info} />;
		});
		var queue;
		if (this.props.queuer) {
			queue = (
				<div>
					<button onClick={this.play}>play</button>
					<button onClick={this.add}>add</button>
				</div>
			);
		};
		return (
			<div>
				{queue}
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

			</div>
		);
	}
});

var TrackList = React.createClass({
	mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
	getInitialState: function() {
		return Stores.tracks.data || {};
	},
	render: function() {
		return <Tracks tracks={this.state.Tracks} queuer={true} />;
	}
});

function searchClass(field) {
	return React.createClass({
		mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
		render: function() {
			if (!Stores.tracks.data) {
				return null;
			}
			var tracks = [];
			var prop = this.props.params[field];
			_.each(Stores.tracks.data.Tracks, function(val) {
				if (val.Info[field] == prop) {
					tracks.push(val);
				}
			});
			return <Tracks tracks={tracks} queuer={true} />;
		}
	});
}

var Artist = searchClass('Artist');
var Album = searchClass('Album');
