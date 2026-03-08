export namespace main {
	
	export class AkipControls {
	    address: string;
	    timeBase: number;
	    hOffset: string;
	    reper: string;
	    square: string;
	    minY: string;
	    minMove: string;
	    autoSearch: boolean;
	    cursorMode: string;
	    cursorPos: number[];
	
	    static createFrom(source: any = {}) {
	        return new AkipControls(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	        this.timeBase = source["timeBase"];
	        this.hOffset = source["hOffset"];
	        this.reper = source["reper"];
	        this.square = source["square"];
	        this.minY = source["minY"];
	        this.minMove = source["minMove"];
	        this.autoSearch = source["autoSearch"];
	        this.cursorMode = source["cursorMode"];
	        this.cursorPos = source["cursorPos"];
	    }
	}
	export class AkipSnapshot {
	    connected: boolean;
	    lastResponse: string;
	    address: string;
	    timeBase: number;
	    hOffset: string;
	    reper: string;
	    square: string;
	    minY: string;
	    minMove: string;
	    autoSearch: boolean;
	    cursorMode: string;
	    cursorPos: number[];
	    x: number[];
	    y: number[];
	    vSpeed: string;
	    vTime: string;
	    volume: string;
	    registration: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AkipSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.lastResponse = source["lastResponse"];
	        this.address = source["address"];
	        this.timeBase = source["timeBase"];
	        this.hOffset = source["hOffset"];
	        this.reper = source["reper"];
	        this.square = source["square"];
	        this.minY = source["minY"];
	        this.minMove = source["minMove"];
	        this.autoSearch = source["autoSearch"];
	        this.cursorMode = source["cursorMode"];
	        this.cursorPos = source["cursorPos"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.vSpeed = source["vSpeed"];
	        this.vTime = source["vTime"];
	        this.volume = source["volume"];
	        this.registration = source["registration"];
	    }
	}
	export class LogEntry {
	    time: string;
	    level: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.time = source["time"];
	        this.level = source["level"];
	        this.message = source["message"];
	    }
	}

}

