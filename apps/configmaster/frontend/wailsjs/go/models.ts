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
	export class MasterSettings {
	    project_name: string;
	    manifest_url: string;
	    s3_endpoint: string;
	    s3_bucket: string;
	    s3_region: string;
	    s3_tenant_id: string;
	    s3_access_key_id: string;
	    s3_secret_key: string;
	    s3_prefix: string;
	    s3_public_base_url: string;
	    uploader_config_path: string;
	    updater_config_path: string;
	    apps: ManagedAppConfig[];
	
	    static createFrom(source: any = {}) {
	        return new MasterSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_name = source["project_name"];
	        this.manifest_url = source["manifest_url"];
	        this.s3_endpoint = source["s3_endpoint"];
	        this.s3_bucket = source["s3_bucket"];
	        this.s3_region = source["s3_region"];
	        this.s3_tenant_id = source["s3_tenant_id"];
	        this.s3_access_key_id = source["s3_access_key_id"];
	        this.s3_secret_key = source["s3_secret_key"];
	        this.s3_prefix = source["s3_prefix"];
	        this.s3_public_base_url = source["s3_public_base_url"];
	        this.uploader_config_path = source["uploader_config_path"];
	        this.updater_config_path = source["updater_config_path"];
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
	export class MasterSnapshot {
	    config_path: string;
	    ready: boolean;
	    error: string;
	    settings: MasterSettings;
	    logs: LogEntry[];
	
	    static createFrom(source: any = {}) {
	        return new MasterSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config_path = source["config_path"];
	        this.ready = source["ready"];
	        this.error = source["error"];
	        this.settings = this.convertValues(source["settings"], MasterSettings);
	        this.logs = this.convertValues(source["logs"], LogEntry);
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

