// @flow

var exports = module.exports = {};

var Mog = require('./mog.js');
var React = require('react');
var Reflux = require('reflux');
var _ = require('underscore');
var mui = require('material-ui');

var { DropDownMenu, FlatButton, IconButton, RaisedButton, TextField } = mui;

exports.Protocols = React.createClass({
	mixins: [Reflux.listenTo(Stores.protocols, 'setState')],
	getInitialState: function() {
		var d = {
			Available: {},
			Current: {},
			Selected: 'file',
		};
		return _.extend(d, Stores.protocols.data);
	},
	handleChange: function(e, idx, item) {
		this.setState({
			Selected: item.text,
			SelectedIndex: idx,
		});
	},
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return { text: protocol };
		});
		var dropdown;
		if (options.length) {
			dropdown = <DropDownMenu onChange={this.handleChange} selectedIndex={this.state.SelectedIndex} menuItems={options} />
		}
		var protocols = [];
		_.each(this.state.Current, function(instances, protocol) {
			_.each(instances, function(key) {
				protocols.push(<ProtocolRow key={key} protocol={protocol} name={key} />);
			}, this);
		}, this);
		var selected;
		if (this.state.Selected) {
			selected = <Protocol protocol={this.state.Selected} params={this.state.Available[this.state.Selected]} />;
		}
		return <div>
			<h2>New Protocol</h2>
			{dropdown}
			{selected}
			<h2>Existing Protocols</h2>
			<table>
				<thead>
					<tr>
						<th>protocol</th>
						<th>name</th>
						<th>remove</th>
						<th>refresh</th>
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
		var params = this.state.params.map(function(v) {
			return {
				name: 'params',
				value: v,
			};
		});
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		Mog.POST('/api/protocol/add', params, function() {
			this.setState(this.getInitialState());
		}.bind(this));
	},
	render: function() {
		if (!this.props.params) {
			return <div></div>;
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
				return <TextField key={idx} style={{width: '75%'}} onChange={change} value={this.state.params[idx]} floatingLabelText={param} />;
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(<FlatButton
				key='oauth'
				primary={true}
				linkButton={true}
				href={this.props.params.OAuthURL}
				label='connect'
				/>);
		}
		var save;
		if (this.state.save) {
			save = <RaisedButton onClick={this.save} label='save' />;
		}
		return (
			<div>
				{params}
				{save}
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
				<td>{this.props.protocol}</td>
				<td>{this.props.name}</td>
				<td>
					<IconButton onClick={this.remove}>
						<i className="material-icons">clear</i>
					</IconButton>
				</td>
				<td>
					<IconButton onClick={this.refresh}>
						<i className="material-icons">refresh</i>
					</IconButton>
				</td>
			</tr>
		);
	}
});