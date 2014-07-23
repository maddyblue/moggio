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
		$.get('/api/list', function(result) {
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

var Protocol = React.createClass({displayName: 'Protocol',
	render: function() {
		var that = this;
		var params = this.props.params.map(function(param, idx) {
			var current = that.props.current || {};
			return React.DOM.li({key: param}, param, ": ", current[idx]);
		});
		return (
			React.DOM.div({key: this.props.key}, 
				React.DOM.h2(null, this.props.key), 
				React.DOM.ul(null, params)
			)
		);
	}
});

var Protocols = React.createClass({displayName: 'Protocols',
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
		var that = this;
		var protocols = keys.map(function(protocol) {
			return Protocol({key: protocol, params: that.state.available[protocol], current: that.state.current[protocol]});
		});
		return React.DOM.div(null, protocols);
	}
});

var routes = {};

var Link = React.createClass({displayName: 'Link',
	componentDidMount: function() {
		routes[this.props.href] = this.props.handler;
		if (this.props.index) {
			routes['/'] = this.props.handler;
		}
	},
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
			React.DOM.ul(null, 
				Link({href: "/list", name: "List", handler: TrackList, index: true}), 
				Link({href: "/protocols", name: "Protocols", handler: Protocols})
			)
		);
	}
});

React.renderComponent(Navigation(null), document.getElementById('navigation'));

function router() {
	var component = routes[window.location.pathname];
	if (!component) {
		alert('unknown route');
	} else {
		React.renderComponent(component(), document.getElementById('main'));
	}
}
router();