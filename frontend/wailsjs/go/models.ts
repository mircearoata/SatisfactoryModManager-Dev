export namespace bindings {
	
	export class Update {
	    item: string;
	    currentVersion: string;
	    newVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new Update(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = source['item'];
	        this.currentVersion = source['currentVersion'];
	        this.newVersion = source['newVersion'];
	    }
	}
	export class InstallationInfo {
	    // Go type: cli.Installation
	    installation?: any;
	    // Go type: types.Installation
	    info?: any;
	
	    static createFrom(source: any = {}) {
	        return new InstallationInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.installation = this.convertValues(source['installation'], null);
	        this.info = this.convertValues(source['info'], null);
	    }
	
	  convertValues(a: any, classs: any, asMap = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ('object' === typeof a) {
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
	export class Progress {
	    item: string;
	    message: string;
	    progress: number;
	
	    static createFrom(source: any = {}) {
	        return new Progress(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.item = source['item'];
	        this.message = source['message'];
	        this.progress = source['progress'];
	    }
	}
	export class SavedModFilters {
	    order: string;
	    filter: string;
	
	    static createFrom(source: any = {}) {
	        return new SavedModFilters(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.order = source['order'];
	        this.filter = source['filter'];
	    }
	}

}

export namespace cli {
	
	export class ProfileMod {
	    version: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileMod(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source['version'];
	        this.enabled = source['enabled'];
	    }
	}
	export class Profile {
	    name: string;
	    mods: {[key: string]: ProfileMod};
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source['name'];
	        this.mods = this.convertValues(source['mods'], ProfileMod, true);
	    }
	
	  convertValues(a: any, classs: any, asMap = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ('object' === typeof a) {
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

export namespace install_finders {
	
	export class Installation {
	    path: string;
	    version: number;
	    branch: string;
	    launcher: string;
	    launchPath: string;
	
	    static createFrom(source: any = {}) {
	        return new Installation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source['path'];
	        this.version = source['version'];
	        this.branch = source['branch'];
	        this.launcher = source['launcher'];
	        this.launchPath = source['launchPath'];
	    }
	}

}