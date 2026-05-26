export namespace main {
	
	export class DBResult {
	    path: string;
	    appName: string;
	    backupFile: string;
	
	    static createFrom(source: any = {}) {
	        return new DBResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.appName = source["appName"];
	        this.backupFile = source["backupFile"];
	    }
	}
	export class FixOptions {
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new FixOptions(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.mode = source["mode"];
	    }
	}
	export class FixResult {
	    success: boolean;
	    total: number;
	    workspaceMapped: number;
	    timestampsInjected: number;
	    dbResults: DBResult[];
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new FixResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.success = source["success"];
	        this.total = source["total"];
	        this.workspaceMapped = source["workspaceMapped"];
	        this.timestampsInjected = source["timestampsInjected"];
	        this.dbResults = this.convertValues(source["dbResults"], DBResult);
	        this.error = source["error"];
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
	export class ResolvedConversation {
	    index: number;
	    id: string;
	    title: string;
	    source: string;
	    hasWorkspace: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ResolvedConversation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.id = source["id"];
	        this.title = source["title"];
	        this.source = source["source"];
	        this.hasWorkspace = source["hasWorkspace"];
	    }
	}
	export class ScanResult {
	    conversations: ResolvedConversation[];
	    stats: Record<string, number>;
	    dirSummary: Record<string, number>;
	    totalExisting: number;
	    wsCount: number;
	    knownWSCount: number;
	
	    static createFrom(source: any = {}) {
	        return new ScanResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.conversations = this.convertValues(source["conversations"], ResolvedConversation);
	        this.stats = source["stats"];
	        this.dirSummary = source["dirSummary"];
	        this.totalExisting = source["totalExisting"];
	        this.wsCount = source["wsCount"];
	        this.knownWSCount = source["knownWSCount"];
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
	export class SystemInfo {
	    os: string;
	    dbPaths: string[];
	    conversationDirs: string[];
	    brainDirs: string[];
	    workspaceStorageDir: string;
	
	    static createFrom(source: any = {}) {
	        return new SystemInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.os = source["os"];
	        this.dbPaths = source["dbPaths"];
	        this.conversationDirs = source["conversationDirs"];
	        this.brainDirs = source["brainDirs"];
	        this.workspaceStorageDir = source["workspaceStorageDir"];
	    }
	}

}

export namespace updater {
	
	export class UpdateInfo {
	    available: boolean;
	    currentVersion: string;
	    latestVersion: string;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.available = source["available"];
	        this.currentVersion = source["currentVersion"];
	        this.latestVersion = source["latestVersion"];
	        this.url = source["url"];
	    }
	}

}

