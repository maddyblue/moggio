var exports = (module.exports = {});

var FixedDataTable = require('fixed-data-table');
var Moggio = require('./moggio.js');
var React = require('react');
var Reflux = require('reflux');
var Router = require('react-router');
var _ = require('underscore');

var Link = Router.Link;

function group(route, field, name) {
	return React.createClass({
		mixins: [Reflux.listenTo(Stores.tracks, 'setState')],
		getInitialState: function() {
			return Stores.tracks.data || {};
		},
		render: function() {
			var entries = {};
			_.each(this.state.Tracks, function(val) {
				var f = val.Info[field];
				if (f) {
					entries[f] = true;
				}
			});
			var list = _.keys(entries);
			list.sort(function(a, b) {
				return a.toLowerCase().localeCompare(b.toLowerCase());
			});
			var lis = _.map(list, function(val) {
				var params = {};
				params[field] = val;
				return (
					<li key={val}>
						<Link to={route} params={params}>
							{val}
						</Link>
					</li>
				);
			});
			return (
				<div>
					<div className="mdl-typography--display-3 mdl-color-text--grey-600">
						<Link to="app">Music</Link> &gt; {name}
					</div>
					<ul>{lis}</ul>
				</div>
			);
		},
	});
}

exports.Artists = group('artist', 'Artist', 'Artists');
exports.Albums = group('album', 'Album', 'Albums');
