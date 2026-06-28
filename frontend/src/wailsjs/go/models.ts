export namespace app {
	
	export class ConfigDTO {
	    url: string;
	    quality: string;
	    format: string;
	    audioOnly: boolean;
	    audioFormat: string;
	    outputPath: string;
	    outputTemplate: string;
	    useCookiesFile: boolean;
	    cookiesFile: string;
	    useCookiesBrowser: boolean;
	    cookiesBrowser: string;
	    proxy: string;
	    rateLimit: string;
	    playlistReverse: boolean;
	    continue: boolean;
	    noPart: boolean;
	    playlistStart: number;
	    playlistEnd: number;
	    maxDownloads: number;
	    downloadArchive: string;
	    noPlaylist: boolean;
	    flatPlaylist: boolean;
	    writeSubs: boolean;
	    writeAutoSub: boolean;
	    embedSubs: boolean;
	    subLangs: string;
	    writeThumbnail: boolean;
	    embedThumbnail: boolean;
	    embedMetadata: boolean;
	    embedChapters: boolean;
	    retries: number;
	    fragmentRetries: number;
	    concurrentFragments: number;
	    socketTimeout: number;
	    noWarnings: boolean;
	    verbose: boolean;
	    quiet: boolean;
	    writeInfoJSON: boolean;
	    loadInfoJson: string;
	    windowsFilenames: boolean;
	    noMtime: boolean;
	    abortOnError: boolean;
	    ignoreErrors: boolean;
	    extraArgs: string;
	    ffmpegLocation: string;
	    username: string;
	    password: string;
	    sponsorBlockRemove: boolean;
	    batchUrls: string;
	    ytDlpPath: string;
	    denoPath: string;
	
	    static createFrom(source: any = {}) {
	        return new ConfigDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.quality = source["quality"];
	        this.format = source["format"];
	        this.audioOnly = source["audioOnly"];
	        this.audioFormat = source["audioFormat"];
	        this.outputPath = source["outputPath"];
	        this.outputTemplate = source["outputTemplate"];
	        this.useCookiesFile = source["useCookiesFile"];
	        this.cookiesFile = source["cookiesFile"];
	        this.useCookiesBrowser = source["useCookiesBrowser"];
	        this.cookiesBrowser = source["cookiesBrowser"];
	        this.proxy = source["proxy"];
	        this.rateLimit = source["rateLimit"];
	        this.playlistReverse = source["playlistReverse"];
	        this.continue = source["continue"];
	        this.noPart = source["noPart"];
	        this.playlistStart = source["playlistStart"];
	        this.playlistEnd = source["playlistEnd"];
	        this.maxDownloads = source["maxDownloads"];
	        this.downloadArchive = source["downloadArchive"];
	        this.noPlaylist = source["noPlaylist"];
	        this.flatPlaylist = source["flatPlaylist"];
	        this.writeSubs = source["writeSubs"];
	        this.writeAutoSub = source["writeAutoSub"];
	        this.embedSubs = source["embedSubs"];
	        this.subLangs = source["subLangs"];
	        this.writeThumbnail = source["writeThumbnail"];
	        this.embedThumbnail = source["embedThumbnail"];
	        this.embedMetadata = source["embedMetadata"];
	        this.embedChapters = source["embedChapters"];
	        this.retries = source["retries"];
	        this.fragmentRetries = source["fragmentRetries"];
	        this.concurrentFragments = source["concurrentFragments"];
	        this.socketTimeout = source["socketTimeout"];
	        this.noWarnings = source["noWarnings"];
	        this.verbose = source["verbose"];
	        this.quiet = source["quiet"];
	        this.writeInfoJSON = source["writeInfoJSON"];
	        this.loadInfoJson = source["loadInfoJson"];
	        this.windowsFilenames = source["windowsFilenames"];
	        this.noMtime = source["noMtime"];
	        this.abortOnError = source["abortOnError"];
	        this.ignoreErrors = source["ignoreErrors"];
	        this.extraArgs = source["extraArgs"];
	        this.ffmpegLocation = source["ffmpegLocation"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.sponsorBlockRemove = source["sponsorBlockRemove"];
	        this.batchUrls = source["batchUrls"];
	        this.ytDlpPath = source["ytDlpPath"];
	        this.denoPath = source["denoPath"];
	    }
	}
	export class AppSettingsDTO {
	    config: ConfigDTO;
	    ffmpegPath: string;
	    sessionPath: string;
	    queueParallel: number;
	    language: string;
	    ytDlpPath: string;
	    denoPath: string;
	    lastProfile: string;
	    debugLog: boolean;
	    activityPanelOpen: boolean;
	    uiScale: number;
	    theme: string;
	    windowWidth: number;
	    windowHeight: number;
	    activityPanelOffset: number;
	
	    static createFrom(source: any = {}) {
	        return new AppSettingsDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.config = this.convertValues(source["config"], ConfigDTO);
	        this.ffmpegPath = source["ffmpegPath"];
	        this.sessionPath = source["sessionPath"];
	        this.queueParallel = source["queueParallel"];
	        this.language = source["language"];
	        this.ytDlpPath = source["ytDlpPath"];
	        this.denoPath = source["denoPath"];
	        this.lastProfile = source["lastProfile"];
	        this.debugLog = source["debugLog"];
	        this.activityPanelOpen = source["activityPanelOpen"];
	        this.uiScale = source["uiScale"];
	        this.theme = source["theme"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.activityPanelOffset = source["activityPanelOffset"];
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
	export class QueueTaskDTO {
	    id: string;
	    name: string;
	    config: ConfigDTO;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueTaskDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.config = this.convertValues(source["config"], ConfigDTO);
	        this.status = source["status"];
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
	export class AppStateDTO {
	    settings: AppSettingsDTO;
	    queue: QueueTaskDTO[];
	    journal: string[];
	    running: boolean;
	    version: string;
	    toolsDir: string;
	    scheduledQueueAt: number;
	
	    static createFrom(source: any = {}) {
	        return new AppStateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.settings = this.convertValues(source["settings"], AppSettingsDTO);
	        this.queue = this.convertValues(source["queue"], QueueTaskDTO);
	        this.journal = source["journal"];
	        this.running = source["running"];
	        this.version = source["version"];
	        this.toolsDir = source["toolsDir"];
	        this.scheduledQueueAt = source["scheduledQueueAt"];
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
	export class AppUpdateDTO {
	    current: string;
	    latest: string;
	    updateAvail: boolean;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new AppUpdateDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.updateAvail = source["updateAvail"];
	        this.url = source["url"];
	    }
	}
	
	export class DependencyDTO {
	    id: string;
	    level: string;
	    status: string;
	    path: string;
	    version: string;
	    latestVer: string;
	    updateAvail: boolean;
	    source: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new DependencyDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.level = source["level"];
	        this.status = source["status"];
	        this.path = source["path"];
	        this.version = source["version"];
	        this.latestVer = source["latestVer"];
	        this.updateAvail = source["updateAvail"];
	        this.source = source["source"];
	        this.error = source["error"];
	    }
	}
	export class HealthIssueDTO {
	    id: string;
	    severity: string;
	    key: string;
	    params: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new HealthIssueDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.severity = source["severity"];
	        this.key = source["key"];
	        this.params = source["params"];
	    }
	}
	export class HistoryItemDTO {
	    at: string;
	    url: string;
	    status: string;
	    output: string;
	    error: string;
	    durationSec: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItemDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.at = source["at"];
	        this.url = source["url"];
	        this.status = source["status"];
	        this.output = source["output"];
	        this.error = source["error"];
	        this.durationSec = source["durationSec"];
	    }
	}
	export class InstallGuardDTO {
	    alreadyInstalled: boolean;
	    foundElsewhere: boolean;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new InstallGuardDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alreadyInstalled = source["alreadyInstalled"];
	        this.foundElsewhere = source["foundElsewhere"];
	        this.path = source["path"];
	    }
	}
	export class InstanceDTO {
	    pid: number;
	    startedAt: string;
	    busy: boolean;
	
	    static createFrom(source: any = {}) {
	        return new InstanceDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.pid = source["pid"];
	        this.startedAt = source["startedAt"];
	        this.busy = source["busy"];
	    }
	}
	export class ProfileDTO {
	    name: string;
	    description: string;
	    config: ConfigDTO;
	
	    static createFrom(source: any = {}) {
	        return new ProfileDTO(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.config = this.convertValues(source["config"], ConfigDTO);
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

export namespace downloader {
	
	export class Format {
	    ID: string;
	    Ext: string;
	    Resolution: string;
	    FPS: number;
	    Vcodec: string;
	    Acodec: string;
	    Filesize: number;
	    TBR: number;
	    Note: string;
	
	    static createFrom(source: any = {}) {
	        return new Format(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ID = source["ID"];
	        this.Ext = source["Ext"];
	        this.Resolution = source["Resolution"];
	        this.FPS = source["FPS"];
	        this.Vcodec = source["Vcodec"];
	        this.Acodec = source["Acodec"];
	        this.Filesize = source["Filesize"];
	        this.TBR = source["TBR"];
	        this.Note = source["Note"];
	    }
	}
	export class MediaEntry {
	    Title: string;
	    ID: string;
	    URL: string;
	    Duration: string;
	    Uploader: string;
	    Thumbnail: string;
	    Formats: Format[];
	
	    static createFrom(source: any = {}) {
	        return new MediaEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.ID = source["ID"];
	        this.URL = source["URL"];
	        this.Duration = source["Duration"];
	        this.Uploader = source["Uploader"];
	        this.Thumbnail = source["Thumbnail"];
	        this.Formats = this.convertValues(source["Formats"], Format);
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
	export class ProbeResult {
	    Title: string;
	    Kind: string;
	    Entries: MediaEntry[];
	    Selected: number;
	
	    static createFrom(source: any = {}) {
	        return new ProbeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Title = source["Title"];
	        this.Kind = source["Kind"];
	        this.Entries = this.convertValues(source["Entries"], MediaEntry);
	        this.Selected = source["Selected"];
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

