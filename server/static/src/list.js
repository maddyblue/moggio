// @flow

var Table = FixedDataTable.Table;
var Column = FixedDataTable.Column;

var Tracks = React.createClass({
	getDefaultProps: function() {
		return {
			tracks: []
		};
	},
	getInitialState: function() {
		var init = {
			sort: this.props.initSort || 'Title',
			asc: true,
			tracks: [],
			search: '',
		};
		if (this.props.isqueue || this.props.useIdxAsNum) {
			init.sort = 'Track';
		}
		return init;
	},
	componentWillReceiveProps: function(next) {
		this.update(null, next.tracks);
	},
	mkparams: function() {
		return _.map(this.state.tracks, function(t, i) {
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
	playTrack: function(index) {
		return function() {
			if (this.props.isqueue) {
				var idx = this.getter(index).idx - 1;
				POST('/api/cmd/play_idx?idx=' + idx);
			} else {
				var params = [
					'clear',
					'add-' + this.getter(index).ID.UID
				];
				POST('/api/queue/change', mkcmd(params), function() {
					POST('/api/cmd/play');
				});
			}
		}.bind(this);
	},
	appendTrack: function(index) {
		return function() {
			var params;
			if (this.props.isqueue) {
				var idx = this.getter(index).idx - 1;
				params = [
					'rem-' + idx
				];
			} else {
				params = [
					'add-' + this.getter(index).ID.UID
				];
			}
			POST('/api/queue/change', mkcmd(params));
		}.bind(this);
	},
	sort: function(field) {
		return function() {
			if (this.state.sort == field) {
				this.update({asc: !this.state.asc});
			} else {
				this.update({sort: field});
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
	handleResize: function() {
		this.forceUpdate();
	},
	componentDidMount: function() {
		window.addEventListener('resize', this.handleResize);
		this.update();
	},
	componentWillUnmount: function() {
		window.removeEventListener('resize', this.handleResize);
	},
	update: function(obj, next) {
		if (obj) {
			this.setState(obj);
		}
		obj = _.extend({}, this.state, obj);
		var tracks = next || this.props.tracks;
		if (next) {
			_.each(tracks, function(v, i) {
				v.idx = i + 1;
			});
		}
		if (obj.search) {
			var s = obj.search.toLocaleLowerCase().trim();
			tracks = _.filter(tracks, function(v) {
				var t = v.Info.Title + v.Info.Album + v.Info.Artist;
				t = t.toLocaleLowerCase();
				return t.indexOf(s) > -1;
			});
		}
		var useIdx = (obj.sort == 'Track' && this.props.useIdxAsNum) || this.props.isqueue;
		tracks = _.sortBy(tracks, function(v) {
			return v.Info.Track;
		});
		tracks = _.sortBy(tracks, function(v) {
			if (useIdx) {
				return v.idx;
			}
			var d = v.Info[obj.sort];
			if (_.isString(d)) {
				d = d.toLocaleLowerCase();
			}
			return d;
		}.bind(this));
		if (!obj.asc) {
			tracks.reverse();
		}
		this.setState({tracks: tracks});
	},
	search: function(event) {
		this.update({search: event.target.value});
	},
	getter: function(index) {
		return this.state.tracks[index];
	},
	timeCellRenderer: function(str, key, data, index) {
		return <div><Time time={data.Info.Time} /></div>;
	},
	timeHeader: function() {
		return function() {
			return <i className={"fa fa-clock-o " + this.sortClass('Time')} onClick={this.sort('Time')} />;
		}.bind(this);
	},
	mkHeader: function(name, text) {
		if (!text) {
			text = name;
		}
		if (this.props.isqueue) {
			return function() {
				return text;
			};
		}
		return function() {
			return <div className={this.sortClass(name)} onClick={this.sort(name)}>{text}</div>;
		}.bind(this);
	},
	trackRenderer: function(str, key, data, index) {
		var track = data.Info.Track || '';
		if (this.props.useIdxAsNum) {
			track = data.idx;
		} else if (this.props.noIdx) {
			track = '';
		}
		return (
			<div>
				<span className="nohover">{track}</span>
				<span className="hover"><i className={mkIcon('fa-play')} onClick={this.playTrack(index)} /></span>
			</div>
		);
	},
	titleCellRenderer: function(str, key, data, index) {
		return (
			<div>
				{data.Info.Title}
				<span className="hover pull-right"><i className={mkIcon(this.props.isqueue ? 'fa-times' : 'fa-plus')} onClick={this.appendTrack(index)} /></span>
			</div>
		);
	},
	artistCellRenderer: function(str, key, data, index) {
		return <div><Link to="artist" params={data.Info}>{data.Info.Artist}</Link></div>;
	},
	albumCellRenderer: function(str, key, data, index) {
		return <div><Link to="album" params={data.Info}>{data.Info.Album}</Link></div>;
	},
	render: function() {
		var height = 0;
		if (this.refs.table) {
			var d = this.refs.table.getDOMNode();
			height = window.innerHeight - d.offsetTop - 62;
		}
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
		var tableWidth = window.innerWidth - 227;
		return (
			<div>
				{queue}
				<div><input type="search" style={{width: tableWidth - 2}} placeholder="search" onChange={this.search} value={this.state.search} /></div>
				<Table ref="table"
					headerHeight={50}
					rowHeight={50}
					rowGetter={this.getter}
					rowsCount={this.state.tracks.length}
					width={tableWidth}
					height={height}
					overflowX={'hidden'}
					>
					<Column
						width={50}
						headerRenderer={this.mkHeader('Track', '#')}
						cellRenderer={this.trackRenderer}
					/>
					<Column
						width={200}
						flexGrow={3}
						cellClassName="nowrap"
						headerRenderer={this.mkHeader('Title')}
						cellRenderer={this.titleCellRenderer}
					/>
					<Column
						width={50}
						cellRenderer={this.timeCellRenderer}
						headerRenderer={this.timeHeader()}
					/>
					<Column
						width={100}
						flexGrow={1}
						cellRenderer={this.artistCellRenderer}
						cellClassName="nowrap"
						headerRenderer={this.mkHeader('Artist')}
					/>
					<Column
						width={100}
						flexGrow={1}
						cellRenderer={this.albumCellRenderer}
						cellClassName="nowrap"
						headerRenderer={this.mkHeader('Album')}
					/>
				</Table>
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
		return (
			<div>
				<h4>Music</h4>
				<Tracks tracks={this.state.Tracks} />
			</div>
		);
	}
});

function searchClass(field, sort) {
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
			return <Tracks tracks={tracks} initSort={sort} />;
		}
	});
}

var Artist = searchClass('Artist', 'Album');
var Album = searchClass('Album', 'Track');
