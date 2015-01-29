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
				protocols.push(<Protocol key={key} protocol={protocol} name={key} />);
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

var ProtocolParam = React.createClass({
	getInitialState: function() {
		return {
			value: '',
			changed: false,
		};
	},
	componentWillReceiveProps: function(props) {
		if (this.state.changed) {
			return;
		}
		this.setState({
			value: props.value,
			changed: true,
		});
	},
	paramChange: function(event) {
		this.setState({
			value: event.target.value,
		});
		this.props.change();
	},
	render: function() {
		return (
			<li>
				{this.props.name} <input type="text" onChange={this.paramChange} value={this.state.value || this.props.value} />
			</li>
		);
	}
});

var ProtocolOAuth = React.createClass({
	render: function() {
		return <li><a href={this.props.url}>connect</a></li>;
	}
});

var Protocol = React.createClass({
	getInitialState: function() {
		return {
			save: false,
		};
	},
	getDefaultProps: function() {
		return {
			instance: {},
			params: {},
		};
	},
	setSave: function() {
		this.setState({save: true});
	},
	save: function() {
		var params = Object.keys(this.refs).sort();
		params = params.map(function(ref) {
			var v = this.refs[ref].state.value;
			this.refs[ref].state.value = '';
			return {
				name: 'params',
				value: v,
			};
		}, this);
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		POST('/api/protocol/add', params, function() {
				this.setState({save: false});
			}.bind(this));
	},
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
		var params = [];
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var current = this.props.instance.Params || [];
				return <ProtocolParam key={param} name={param} ref={idx} value={current[idx]} change={this.setSave} />;
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(<ProtocolOAuth key={'oauth'} url={this.props.params.OAuthURL} />);
		}
		if (this.props.name) {
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
		var save;
		if (this.state.save) {
			save = <button onClick={this.save}>save</button>;
		}
		return <div>
				<ul>{params}</ul>
				{save}
			</div>;
	}
});
