// @flow

var Protocols = React.createClass({
	mixins: [Reflux.listenTo(Stores.protocols, 'setState')],
	getInitialState: function() {
		var d = {
			Available: {},
			Current: {},
			Selected: 'file',
		};
		return _.extend(d, Stores.protocols.data);
	},
	handleChange: function(event) {
		this.setState({Selected: event.target.value});
	},
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return <option key={protocol}>{protocol}</option>;
		}.bind(this));
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
			<select onChange={this.handleChange} value={this.state.Selected}>{options}</select>
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
		POST('/api/protocol/add', params, function() {
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
				return (
					<li key={idx}>
						{param}: <input type="text" style={{width: '75%'}} onChange={change} value={this.state.params[idx]} />
					</li>
				);
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(<li key={'oauth'}><a href={this.props.params.OAuthURL}>connect</a></li>);
		}
		var save;
		if (this.state.save) {
			save = <button onClick={this.save}>save</button>;
		}
		return (
			<div>
				<ul>{params}</ul>
				{save}
			</div>
		);
	}
});

var ProtocolRow = React.createClass({
	remove: function() {
		POST('/api/protocol/remove', {
			protocol: this.props.protocol,
			key: this.props.name,
		});
	},
	refresh: function() {
		POST('/api/protocol/refresh', {
			protocol: this.props.protocol,
			key: this.props.name,
		});
	},
	render: function() {
		var icon = 'fa fa-fw fa-border fa-2x clickable ';
		return (
			<tr>
				<td>{this.props.protocol}</td>
				<td>{this.props.name}</td>
				<td><i className={icon + 'fa-times'} onClick={this.remove} /></td>
				<td><i className={icon + 'fa-repeat'} onClick={this.refresh} /></td>
			</tr>
		);
	}
});