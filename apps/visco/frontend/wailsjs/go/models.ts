export namespace main {
	
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
	export class ViskoControls {
	    address: string;
	
	    static createFrom(source: any = {}) {
	        return new ViskoControls(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.address = source["address"];
	    }
	}
	export class ViskoRow {
	    t1: number;
	    t2: number;
	    u1: number;
	    u2: number;
	    temp: number;
	
	    static createFrom(source: any = {}) {
	        return new ViskoRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.t1 = source["t1"];
	        this.t2 = source["t2"];
	        this.u1 = source["u1"];
	        this.u2 = source["u2"];
	        this.temp = source["temp"];
	    }
	}
	export class ViskoSnapshot {
	    connected: boolean;
	    lastResponse: string;
	    address: string;
	    rows: ViskoRow[];
	    cursorIndex: number;
	    curT1: string;
	    curT2: string;
	    curU1: string;
	    curU2: string;
	    curTemp: string;
	    curCmd: string;
	    selT1: string;
	    selT2: string;
	    selU1: string;
	    selU2: string;
	    selTemp: string;
	
	    static createFrom(source: any = {}) {
	        return new ViskoSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.connected = source["connected"];
	        this.lastResponse = source["lastResponse"];
	        this.address = source["address"];
	        this.rows = this.convertValues(source["rows"], ViskoRow);
	        this.cursorIndex = source["cursorIndex"];
	        this.curT1 = source["curT1"];
	        this.curT2 = source["curT2"];
	        this.curU1 = source["curU1"];
	        this.curU2 = source["curU2"];
	        this.curTemp = source["curTemp"];
	        this.curCmd = source["curCmd"];
	        this.selT1 = source["selT1"];
	        this.selT2 = source["selT2"];
	        this.selU1 = source["selU1"];
	        this.selU2 = source["selU2"];
	        this.selTemp = source["selTemp"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

