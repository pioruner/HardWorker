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
	export class ManagedAppConfig {
	    id: string;
	    label: string;
	    platform: string;
	    arch: string;
	    install_dir: string;
	    executable: string;
	
	    static createFrom(source: any = {}) {
	        return new ManagedAppConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.platform = source["platform"];
	        this.arch = source["arch"];
	        this.install_dir = source["install_dir"];
	        this.executable = source["executable"];
	    }
	}
	export class ManagedAppState {
	    id: string;
	    label: string;
	    platform: string;
	    arch: string;
	    install_dir: string;
	    executable: string;
	    current_version: string;
	    latest_version: string;
	    status: string;
	    message: string;
	    has_update: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ManagedAppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.platform = source["platform"];
	        this.arch = source["arch"];
	        this.install_dir = source["install_dir"];
	        this.executable = source["executable"];
	        this.current_version = source["current_version"];
	        this.latest_version = source["latest_version"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.has_update = source["has_update"];
	    }
	}
	export class SaveSettingsInput {
	    manifest_url: string;
	    s3_endpoint: string;
	    s3_bucket: string;
	    s3_region: string;
	    s3_tenant_id: string;
	    s3_access_key_id: string;
	    s3_secret_key: string;
	    s3_prefix: string;
	    s3_public_base_url: string;
	    selected_app_id: string;
	    apps: ManagedAppConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SaveSettingsInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manifest_url = source["manifest_url"];
	        this.s3_endpoint = source["s3_endpoint"];
	        this.s3_bucket = source["s3_bucket"];
	        this.s3_region = source["s3_region"];
	        this.s3_tenant_id = source["s3_tenant_id"];
	        this.s3_access_key_id = source["s3_access_key_id"];
	        this.s3_secret_key = source["s3_secret_key"];
	        this.s3_prefix = source["s3_prefix"];
	        this.s3_public_base_url = source["s3_public_base_url"];
	        this.selected_app_id = source["selected_app_id"];
	        this.apps = this.convertValues(source["apps"], ManagedAppConfig);
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
	export class SettingsState {
	    manifest_url: string;
	    s3_endpoint: string;
	    s3_bucket: string;
	    s3_region: string;
	    s3_tenant_id: string;
	    s3_access_key_id: string;
	    s3_secret_key: string;
	    s3_prefix: string;
	    s3_public_base_url: string;
	    selected_app_id: string;
	    apps: ManagedAppConfig[];
	
	    static createFrom(source: any = {}) {
	        return new SettingsState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.manifest_url = source["manifest_url"];
	        this.s3_endpoint = source["s3_endpoint"];
	        this.s3_bucket = source["s3_bucket"];
	        this.s3_region = source["s3_region"];
	        this.s3_tenant_id = source["s3_tenant_id"];
	        this.s3_access_key_id = source["s3_access_key_id"];
	        this.s3_secret_key = source["s3_secret_key"];
	        this.s3_prefix = source["s3_prefix"];
	        this.s3_public_base_url = source["s3_public_base_url"];
	        this.selected_app_id = source["selected_app_id"];
	        this.apps = this.convertValues(source["apps"], ManagedAppConfig);
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
	export class UpdaterSnapshot {
	    config_path: string;
	    ready: boolean;
	    busy: boolean;
	    error: string;
	    apps: ManagedAppState[];
	    progress: updater.Progress;
	    logs: LogEntry[];
	    settings: SettingsState;
	
	    static createFrom(source: any = {}) {
	        return new UpdaterSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config_path = source["config_path"];
	        this.ready = source["ready"];
	        this.busy = source["busy"];
	        this.error = source["error"];
	        this.apps = this.convertValues(source["apps"], ManagedAppState);
	        this.progress = this.convertValues(source["progress"], updater.Progress);
	        this.logs = this.convertValues(source["logs"], LogEntry);
	        this.settings = this.convertValues(source["settings"], SettingsState);
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

export namespace updater {
	
	export class Progress {
	    stage: string;
	    message: string;
	    percent: number;
	    bytes_done: number;
	    bytes_total: number;
	    // Go type: time
	    started_at: any;
	    // Go type: time
	    finished_at: any;
	    download_path?: string;
	
	    static createFrom(source: any = {}) {
	        return new Progress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.stage = source["stage"];
	        this.message = source["message"];
	        this.percent = source["percent"];
	        this.bytes_done = source["bytes_done"];
	        this.bytes_total = source["bytes_total"];
	        this.started_at = this.convertValues(source["started_at"], null);
	        this.finished_at = this.convertValues(source["finished_at"], null);
	        this.download_path = source["download_path"];
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

