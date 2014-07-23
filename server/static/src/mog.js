/** @jsx React.DOM */

var TrackListRow = React.createClass({
	render: function() {
		return (<tr><td>{this.props.protocol}</td><td>{this.props.id}</td></tr>);
	}
});

var Track = React.createClass({
	render: function() {
		return (
			<tr>
				<td>{this.props.protocol}</td>
				<td>{this.props.id}</td>
			</tr>
		);
	}
});

var TrackList = React.createClass({
	getInitialState: function() {
		return {
			tracks: []
		};
	},
	componentDidMount: function() {
		$.get('/api/list', function(result) {
			this.setState({tracks: result});
		}.bind(this));
	},
	render: function() {
		var tracks = this.state.tracks.map(function (t) {
			return <Track protocol={t[0]} id={t[1]} key={t[0] + '|' + t[1]} />;
		});
		return (
			<table>
				<tbody>{tracks}</tbody>
			</table>
		);
	}
});

var Protocols = React.createClass({
	getInitialState: function() {
		return {
			available: {},
			current: {},
		};
	},
	componentDidMount: function() {
		$.get('/api/protocol/get', function(result) {
			this.setState({available: result});
		}.bind(this));
		$.get('/api/protocol/list', function(result) {
			this.setState({current: result});
		}.bind(this));
	},
	render: function() {
		var keys = Object.keys(this.state.available);
		keys.sort();
		var protocols = keys.map(function(protocol) {
			return (
				<div key={protocol}>
					<h2>{protocol}</h2>
				</div>
			);
		});
		return <div>{protocols}</div>;
	}
});

var Link = React.createClass({
	click: function(event) {
		history.pushState(null, this.props.Name, this.props.href);
		router();
		event.preventDefault();
	},
	render: function() {
		return <li><a href={this.props.href} onClick={this.click}>{this.props.name}</a></li>
	}
});

var Navigation = React.createClass({
	render: function() {
		return (
			<ul>
				<Link href="/list" name="List" />
				<Link href="/protocols" name="Protocols" />
			</ul>
		);
	}
});

React.renderComponent(<Navigation />, document.getElementById('navigation'));

function router() {
	var component;
	switch (window.location.pathname) {
	case '/':
	case '/list':
		component = <TrackList />;
		break;
	case '/protocols':
		component = <Protocols />;
		break;
	default:
		alert('Unknown route');
		break;
	}
	if (component) {
		React.renderComponent(component, document.getElementById('main'));
	}
}
router();