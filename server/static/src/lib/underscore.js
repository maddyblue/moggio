declare class UnderscoreStatic {
	each(list: Array<any>, iteratee: any, context?: any): void;
	extend(destination: Object, ...sources: Object): Object;
	isArray(object: Object): boolean;
	isObject(object: Object): boolean;
	map(list: Array<any>, iteratee: any, context?: any): Array<any>;
}

declare var _: UnderscoreStatic;