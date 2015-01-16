declare class RefluxStatic {
	createActions(list: Array<string>): Array<any>;
	createStore(store: Object): any;
	listenTo(store: any, s: string): any;
}

declare var Reflux: RefluxStatic;