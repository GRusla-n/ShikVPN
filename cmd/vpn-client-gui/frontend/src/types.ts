export interface StatusUpdate {
  status: 'disconnected' | 'connecting' | 'connected' | 'error';
  assignedIP: string;
  error: string;
}

export interface LogEntry {
  timestamp: string;
  message: string;
}

export interface ClientConfig {
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
}

export type Page = 'connection' | 'config' | 'logs';
