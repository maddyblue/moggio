// @flow

var Track = React.createClass({
	mixins: [Reflux.listenTo(Stores.tracks, 'update')],
	play: function() {
		if (this.props.isqueue) {
			POST('/api/cmd/play_idx?idx=' + this.props.idx);
		} else {
			var params = mkcmd([
				'clear',
				'add-' + this.props.id.UID
			]);
			POST('/api/queue/change', params, function() {
				POST('/api/cmd/play');
			});
		}
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
	over: function() {
		this.setState({over: true});
	},
	out: function() {
		this.setState({over: false});
	},
	dequeue: function() {
		var params = mkcmd([
			'rem-' + this.props.idx
		]);
		POST('/api/queue/change', params);
	},
	append: function() {
		var params = mkcmd([
			'add-' + this.props.id.UID
		]);
		POST('/api/queue/change', params);
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
		var control;
		var track;
		var icon = "fa fa-border fa-lg clickable ";
		if (this.state.over) {
			if (this.props.isqueue) {
				control = <i className={icon + "fa-times"} onClick={this.dequeue} />;
			} else {
				control = <i className={icon + "fa-plus"} onClick={this.append} />;
			}
			track = <i className={icon + "fa-play"} onClick={this.play} />;
		} else {
			track = info.Track || '';
			if (this.props.useIdxAsNum) {
				track = this.props.idx + 1;
			}
		}
		return (
			<tr onMouseEnter={this.over} onMouseLeave={this.out}>
				<td className="control">{track}</td>
				<td>{info.Title}</td>
				<td className="control">{control}</td>
				<td><Time time={info.Time} /></td>
				<td><Link to="artist" params={info}>{info.Artist}</Link></td>
				<td><Link to="album" params={info}>{info.Album}</Link></td>
			</tr>
		);
	}
});

var Tracks = React.createClass({
	getInitialState: function() {
		return {
			sort: 'Title',
			asc: true,
		};
	},
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
	sort: function(field) {
		return function() {
			if (this.state.sort == field) {
				this.setState({asc: !this.state.asc});
			} else {
				this.setState({sort: field});
			}
		}.bind(this);
	},
	sortClass: function(field) {
		if (this.props.isqueue) {
			return '';
		}
		var name = 'clickable ';
		if (this.state.sort == field) {
			name += this.state.asc ? 'sort-asc' : 'sort-desc';
		}
		return name;
	},
	render: function() {
		var sorted = this.props.tracks;
		if (!this.props.isqueue) {
			sorted = _.sortBy(this.props.tracks, function(v) {
				if (!v.Info) {
					return v.ID.UID;
				}
				var d = v.Info[this.state.sort];
				if (_.isString(d)) {
					d = d.toLocaleLowerCase();
				}
				return d;
			}.bind(this));
			if (!this.state.asc) {
				sorted.reverse();
			}
		}
		var tracks = _.map(sorted, function(t, idx) {
			return <Track key={idx + '-' + t.ID.UID} id={t.ID} info={t.Info} idx={idx} isqueue={this.props.isqueue} useIdxAsNum={this.props.useIdxAsNum} />;
		}.bind(this));
		var queue;
		if (!this.props.isqueue) {
			queue = (
				<div>
					<button onClick={this.play}>play</button>
					<button onClick={this.add}>add</button>
				</div>
			);
		};
		var track = this.props.isqueue ? <th></th> : <th className={this.sortClass('Track')} onClick={this.sort('Track')}>#</th>;
		return (
			<div>
				{queue}
				<table className="u-full-width tracks">
					<thead>
						<tr>
							{track}
							<th className={this.sortClass('Title')} onClick={this.sort('Title')}>Name</th>
							<th></th>
							<th className={this.sortClass('Time')} onClick={this.sort('Time')}><i className="fa fa-clock-o" /></th>
							<th className={this.sortClass('Artist')} onClick={this.sort('Artist')}>Artist</th>
							<th className={this.sortClass('Album')} onClick={this.sort('Album')}>Album</th>
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
		return <Tracks tracks={this.state.Tracks} />;
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
			return <Tracks tracks={tracks} />;
		}
	});
}

var Artist = searchClass('Artist');
var Album = searchClass('Album');
