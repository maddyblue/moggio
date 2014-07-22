/** @jsx React.DOM */

var TrackListRow = React.createClass({displayName: 'TrackListRow',
	render: function() {
		return (React.DOM.tr(null, React.DOM.td(null, this.props.protocol), React.DOM.td(null, this.props.id)));
	}
});

var Track = React.createClass({displayName: 'Track',
	render: function() {
		return (
			React.DOM.tr(null, 
				React.DOM.td(null, this.props.protocol), 
				React.DOM.td(null, this.props.id)
			)
		);
	}
});

var TrackList = React.createClass({displayName: 'TrackList',
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
			return Track({protocol: t[0], id: t[1], key: t[0] + '|' + t[1]});
		});
		return (
			React.DOM.table(null, 
				React.DOM.tbody(null, tracks)
			)
		);
	}
});

var Link = React.createClass({displayName: 'Link',
	click: function(event) {
		history.pushState(null, this.props.Name, this.props.href);
		router();
		event.preventDefault();
	},
	render: function() {
		return React.DOM.li(null, React.DOM.a({href: this.props.href, onClick: this.click}, this.props.name))
	}
});

var Navigation = React.createClass({displayName: 'Navigation',
	render: function() {
		return (
			Link({href: "/list", name: "List"})
		);
	}
});

React.renderComponent(Navigation(null), document.getElementById('navigation'));

function router() {
	switch (window.location.pathname) {
	case '/':
	case '/list':
		React.renderComponent(TrackList({source: "/api/list"}), document.getElementById('main'));
		break;
	default:
		alert('Unknown route');
		break;
	}
}
router();