export interface Querier {
  db (key?: string | undefined): string | undefined
  get (key: string): Promise<QueryResult>
  set (key: string, value: string): Promise<QueryResult>
  delete (key: string): Promise<QueryResult>
  keys (): Promise<Array<QueryResult>>
};

export interface QuerierOption {
  user: string;
  repo: string;
  db: string;
  branch: string;
  debug: boolean;
  token: string;
};

export interface QueryResult {
  content?: string;
  name?: string;
  size?: number;
  raw_url?: string;
  html_url?: string;
  commit?: string;
}

