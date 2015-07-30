var exports = module.exports = {};

var Mog = require('./mog.js');
var React = require('react');
var Reflux = require('reflux');
var _ = require('underscore');

var { Button, TextField } = require('./mdl.js');

exports.Protocols = React.createClass({
	mixins: [Reflux.listenTo(Stores.protocols, 'setState')],
	getInitialState: function() {
		var d = {
			Available: {},
			Current: {},
			Selected: 'gmusic',
		};
		return _.extend(d, Stores.protocols.data);
	},
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var selectedIndex = _.indexOf(keys, this.state.Selected);
		var dropdown;
		if (keys.length) {
			var tabs = keys.map(function(protocol) {
				var click = function(evt) {
					evt.preventDefault();
					this.setState({Selected: protocol});
				}.bind(this);
				var cn = 'mdl-tabs__tab';
				if (this.state.Selected == protocol) {
					cn += ' is-active';
				}
				return <a href key={protocol} className={cn} onClick={click}>{protocol}</a>;
			}.bind(this));
			dropdown = (
				<div className="mdl-tabs mdl-js-tabs mdl-js-ripple-effect">
					<div className="mdl-tabs__tab-bar">
						{tabs}
					</div>
				</div>
			);
		}
		var protocols = [];
		_.each(this.state.Current, function(instances, protocol) {
			_.each(instances, function(key) {
				protocols.push(<ProtocolRow key={key} protocol={protocol} name={key} />);
			}, this);
		}, this);
		var selected;
		if (this.state.Selected) {
			selected = (
				<div className="mdl-tabs__panel is-active">
					<Protocol protocol={this.state.Selected} params={this.state.Available[this.state.Selected]} />
				</div>
			);
		}
		return <div>
			<div className="mdl-typography--display-3 mdl-color-text--grey-600">New Protocol</div>
			{dropdown}
			{selected}
			<div className="mdl-typography--display-3 mdl-color-text--grey-600">Existing Protocols</div>
			<table className="mdl-data-table mdl-js-data-table">
				<thead>
					<tr>
						<th className="mdl-data-table__cell--non-numeric">protocol</th>
						<th className="mdl-data-table__cell--non-numeric">name</th>
						<th className="mdl-data-table__cell--non-numeric">remove</th>
						<th className="mdl-data-table__cell--non-numeric">refresh</th>
					</tr>
				</thead>
				<tbody>
					{protocols}
				</tbody>
			</table>
		</div>;
	}
});

var Protocol = React.createClass({
	getInitialState: function() {
		return {
			params: [],
			save: false,
		};
	},
	save: function() {
		params = {
			protocol: this.props.protocol,
			params: this.state.params,
		};
		Mog.POST('/api/protocol/add', params, function() {
			this.setState(this.getInitialState());
		}.bind(this));
	},
	render: function() {
		if (!this.props.params) {
			return <div/>;
		}
		var params = [];
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var change = function(event) {
					var p = this.state.params.slice();
					p[idx] = event.target.value;
					this.setState({
						params: p,
						save: true,
					});
				}.bind(this);
				return <TextField key={idx} style={{width: '75%'}} onChange={change} value={this.state.params[idx]} floating={true} type={param}>{param}</TextField>;
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(
				<Button key="oauth">
					<a href={this.props.params.OAuthURL}>connect</a>
				</Button>
			);
		} else {
			params.push(<Button key="save" raised={true} colored={true} onClick={this.save} disabled={!this.state.save}>save</Button>);
		}
		return (
			<div>
				{params}
			</div>
		);
	}
});

var ProtocolRow = React.createClass({
	remove: function() {
		Mog.POST('/api/protocol/remove', {
			protocol: this.props.protocol,
			key: this.props.name,
		});
	},
	refresh: function() {
		Mog.POST('/api/protocol/refresh', {
			protocol: this.props.protocol,
			key: this.props.name,
		});
	},
	render: function() {
		return (
			<tr>
				<td className="mdl-data-table__cell--non-numeric">{this.props.protocol}</td>
				<td className="mdl-data-table__cell--non-numeric">{this.props.name}</td>
				<td className="mdl-data-table__cell--non-numeric">
					<Button onClick={this.remove} icon={true}>
						<i className="material-icons">clear</i>
					</Button>
				</td>
				<td className="mdl-data-table__cell--non-numeric">
					<Button onClick={this.refresh} icon={true}>
						<i className="material-icons">refresh</i>
					</Button>
				</td>
			</tr>
		);
	}
});