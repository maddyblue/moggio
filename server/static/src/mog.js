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
		$.get(this.props.source, function(result) {
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
			<Link href="/list" name="List" />
		);
	}
});

React.renderComponent(<Navigation />, document.getElementById('navigation'));

function router() {
	switch (window.location.pathname) {
	case '/':
	case '/list':
		React.renderComponent(<TrackList source="/api/list" />, document.getElementById('main'));
		break;
	default:
		alert('Unknown route');
		break;
	}
}
router();