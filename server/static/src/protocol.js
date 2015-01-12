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
	render: function() {
		var keys = Object.keys(this.state.Available) || [];
		keys.sort();
		var options = keys.map(function(protocol) {
			return <option key={protocol}>{protocol}</option>;
		}.bind(this));
		var protocols = [];
		_.each(this.state.Current, function(instances, protocol) {
			_.each(instances, function(inst, key) {
				protocols.push(<Protocol key={'current-' + protocol + '-' + key} protocol={protocol} params={this.state.Available[protocol]} instance={inst} name={key} />);
			}, this);
		}, this);
		var selected;
		if (this.state.Selected) {
			selected = <Protocol key={'selected-' + this.state.Selected} protocol={this.state.Selected} params={this.state.Available[this.state.Selected]} />;
		}
		return <div>
			<h2>New Protocol</h2>
			<select onChange={this.handleChange} value={this.state.Selected}>{options}</select>
			{selected}
			<h2>Existing Protocols</h2>
			{protocols}
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
				{this.props.key} <input type="text" onChange={this.paramChange} value={this.state.value || this.props.value} disabled={this.props.disabled ? 'disabled' : ''} />
			</li>
		);
	}
});

var ProtocolOAuth = React.createClass({
	render: function() {
		var token;
		if (this.props.token) {
			token = <div>Connected until {this.props.token.expiry}</div>;
		}
		return (
			<li>
				{token}
				<a href={this.props.url}>connect</a>
			</li>
		);
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
			return {
				name: 'params',
				value: this.refs[ref].state.value,
			};
		}, this);
		params.push({
			name: 'protocol',
			value: this.props.protocol,
		});
		$.get('/api/protocol/add?' + $.param(params))
			.success(function() {
				this.setState({save: false});
			}.bind(this))
			.error(function(result) {
				console.log(result.responseText);
			});
	},
	remove: function() {
		$.get('/api/protocol/remove?' + $.param({
			protocol: this.props.protocol,
			key: this.props.name,
		}));
	},
	render: function() {
		var params = [];
		var disabled = !!this.props.name;
		if (this.props.params.Params) {
			params = this.props.params.Params.map(function(param, idx) {
				var current = this.props.instance.Params || [];
				return <ProtocolParam key={param} ref={idx} value={current[idx]} change={this.setSave} disabled={disabled} />;
			}.bind(this));
		}
		if (this.props.params.OAuthURL) {
			params.push(<ProtocolOAuth key={'oauth-' + this.props.key} url={this.props.params.OAuthURL} token={this.props.instance.OAuthToken} disabled={disabled} />);
		}
		var save;
		if (this.state.save) {
			save = <button onClick={this.save}>save</button>;
		}
		var title;
		if (this.props.name) {
			title = <h3>{this.props.protocol}: {this.props.name}
					<small><button onClick={this.remove}>remove</button></small>
				</h3>;
		}
		return <div>
				{title}
				<ul>{params}</ul>
				{save}
			</div>;
	}
});
