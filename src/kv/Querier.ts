export interface Committer {
  name: string;
  email: string;
};

export interface Querier {
  db (key?: string | undefined): string | undefined
  get (key: string): Promise<KeyRecord>
  set (key: string, value: string): Promise<KeyRecord>
  delete (key: string): Promise<KeyRecord>
  keys (): Promise<KeyRecord[]>
};

export interface QuerierOption {
  user: string;
  repo: string;
  db: string;
  branch: string;
  debug: boolean;
  token: string;
  committer: Committer
};

export interface KeyRecord {
  // Key content
  content?: string;
  // Key name
  name?: string;
  // Key content size, if the key do not exist, then size = -1
  size?: number;
  // Key git raw url
  raw_url?: string;
  // Key git html url
  html_url?: string;
  // Key git commit hash if there is
  commit?: string;
}

