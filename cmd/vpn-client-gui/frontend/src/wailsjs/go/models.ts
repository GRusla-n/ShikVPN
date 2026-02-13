export namespace config {
	
	export class ClientConfig {
	    server_endpoint: string;
	    server_api_url: string;
	    server_public_key: string;
	    private_key: string;
	    address: string;
	    dns: string;
	    mtu: number;
	    persistent_keepalive: number;
	    interface_name: string;
	    api_key: string;
	    log_level: string;
	
	    static createFrom(source: any = {}) {
	        return new ClientConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.server_endpoint = source["server_endpoint"];
	        this.server_api_url = source["server_api_url"];
	        this.server_public_key = source["server_public_key"];
	        this.private_key = source["private_key"];
	        this.address = source["address"];
	        this.dns = source["dns"];
	        this.mtu = source["mtu"];
	        this.persistent_keepalive = source["persistent_keepalive"];
	        this.interface_name = source["interface_name"];
	        this.api_key = source["api_key"];
	        this.log_level = source["log_level"];
	    }
	}

}

export namespace main {
	
	export class StatusUpdate {
	    status: string;
	    assignedIP: string;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new StatusUpdate(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.status = source["status"];
	        this.assignedIP = source["assignedIP"];
	        this.error = source["error"];
	    }
	}

}

