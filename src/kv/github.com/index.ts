import request from 'request';
import { Logger, createLogger } from '../../helper/log'
import { Querier, QuerierOption, QueryResult } from '../Querier';

interface GithubCommitter {
  name: string;
  email: string;
}

interface GithubQuerierOption extends QuerierOption {
  committer?: GithubCommitter
}

interface GithubAPIGetResult extends QueryResult {
  encoding: string;
  path: string;
  sha: string;
  type: string;
  download_url: string;
  html_url: string;
}

interface GithubQueryOption {
  message: string;
  content?: string;
  sha?: string;
  branch?: string;
  committer?: GithubCommitter
}

interface GithubAPIUpdateResult {
  commit: { sha: string };
  content: GithubAPIGetResult
}

interface PathShaMap {
  [key: string]: string;
}

export default class GithubQuerier implements Querier {

  private readonly option: GithubQuerierOption
  private baseURL: string
  private committer: GithubCommitter
  private shaMap: PathShaMap = {}
  private logger: Logger

  constructor (option: GithubQuerierOption) {
    this.option = option;

    this.committer = <GithubCommitter>{
      name: 'GitDB',
      email: 'GitDB@notavaliable.com',
      ...(this.option.committer || {})
    };

    this.baseURL = `https://api.github.com/repos/${this.option.user}/${this.option.repo}/contents`;

    this.logger = createLogger({ debug: this.option.debug });
  }

  /**
   * Get or set db
   * @param {undefined | string} db
   * @return {undefined | string}
   * @example
   * let db = this.db(); // get current db
   * this.db('test-db'); // set current db to 'test-db'
   */
  db (db?: string): undefined | string {
    if (db) {
      this.option.db = db;
      this.shaMap = {};
    } else {
      return this.option.db;
    }
  }

  /*
   * List all keys in db
   */
  keys (): Promise<QueryResult[]> {
    this.logger.log('github', `List all keys:`);
    return this.query<GithubAPIGetResult[]>().then((res: GithubAPIGetResult[]) => {
      this.logger.log('github', `Get keys count:` + res.length);
      var rs: QueryResult[] = [];
      for (var i: number = 0, l: number = res.length; i < l; i++) {
        var tmp: GithubAPIGetResult = res[i];
        if (tmp.type === 'file') {
          this.shaMap[<string>tmp.name] = tmp.sha;
          rs.push(<QueryResult>{
            name: tmp.name,
            path: tmp.path,
            size: tmp.size,
            html_url: tmp.html_url,
            raw_url: tmp.download_url,
          });
        }
      }
      return rs;
    });
  }

  /*
   * Get key
   */
  get (key: string): Promise<QueryResult> {
    this.logger.log('github', `Get key: ${key}`);
    return this.query<GithubAPIGetResult>(key).then((res) => {
      if (res.size !== -1) {
        this.shaMap[key] = res.sha;
        this.logger.log('github', `Update key sha: ${key}(sha) = ${res.sha}`);
      } else {
        this.logger.log('github', `Get key: ${key} is not found`);
      }
      return <QueryResult>{
        name: res.name,
        content: res.content,
        size: res.size,
        raw_url: res.download_url,
        html_url: res.html_url,
      };
    });
  }

  /*
   * Set key
   */
  set (key: string, value: string): Promise<QueryResult> {
    this.logger.log('github', `Set key: ${key}=${value}`);
    const content: string = Buffer.from(value).toString('base64');
    var op: GithubQueryOption = {
      message: 'GitDB update a key',
      content: content,
    };
    if (this.shaMap[key]) {
      op.sha = this.shaMap[key];
      this.logger.log('github', `Found cached sha: ${key}(sha) = ${op.sha}`);
    } else {
      op.message = 'GitDB create a key';
      this.logger.log('github', `Create new file: ${key}`);
    }
    return this.query<GithubAPIUpdateResult>(key, 'PUT', op).then((res: GithubAPIUpdateResult) => {
      this.shaMap[key] = res.content.sha;
      this.logger.log('github', `Set done. Update key sha: ${key}(sha) = ${res.content.sha}`);
      const qr = <QueryResult>{
        raw_url: res.content.download_url,
        html_url: res.content.html_url,
        size: res.content.size,
        path: res.content.path,
        content: res.content.content || value,
        name: res.content.name,
      };
      if (res.commit) {
        qr.commit = res.commit.sha;
      }
      return qr;
    }).catch(e => {
      if (e.message.indexOf(`"sha" wasn't supplied`) > -1) {
        this.logger.log('github', `Sha wasn't supplied for ${key}, will try to get sha first.`)
        return this.get(key).then(() => {
          return this.set(key, value);
        });
      } else {
        throw e;
      }
    });
  }

  delete (key: string): Promise<QueryResult> {
    this.logger.log('github', `Delete key: ${key}`);
    var op: GithubQueryOption = {
      message: 'GitDB delete a key',
    };
    if (this.shaMap[key]) {
      op.sha = this.shaMap[key];
    }
    return this.query<GithubAPIUpdateResult>(key, 'DELETE', op).then((res: GithubAPIUpdateResult) => {
      return <QueryResult>{
        commit: res.commit ? res.commit.sha : ''
      };
    }).catch(e => {
      if (e.message.indexOf(`"sha" wasn't supplied`) > -1) {
        this.logger.log('github', `Sha wasn't supplied for ${key}, will try to get sha first.`)
        return this.get(key).then((res: QueryResult) => {
          if (res.size === -1) { // Key is not found
            return <QueryResult>{
              commit: ''
            };
          }
          return this.delete(key);
        });
      } else {
        throw e;
      }
    });
  }

  private query<T extends GithubAPIGetResult | GithubAPIUpdateResult | Array<QueryResult>>(key?: string, method?: string, data?: GithubQueryOption): Promise<T> {
    const op: any = {
      url: key ? (this.baseURL + '/' + this.option.db + '/' + key) : (this.baseURL + '/' + this.option.db),
      headers: {
        'User-Agent': 'GitDB',
        'Authorization': 'token ' + this.option.token
      }
    };
    if (method) {
      op.method = method;
    }
    if (data) {
      data.committer = this.committer;
      data.branch = this.option.branch;
      op.body = data;
      op.json = true;
    }
    return new Promise<T>((resolve, reject) => {
      request(op, (err: Error, response: any, body: string | GithubAPIUpdateResult) => {
        if (err) {
          reject(err);
          return;
        }
        if (response.statusCode === 404) {
          if (key) {
            resolve(<T>{
              size: -1,
              content: '',
              name: '',
              type: 'null'
            });
          } else { // keys return;
            resolve(<T><QueryResult[]>[]);
          }
          return;
        } else if (response.statusCode === 201) { // added a file success
          resolve(<T>body);
          return;
        } else if (response.statusCode === 422) {
          reject(body);
          return;
        } else if (response.statusCode !== 200) {
          reject(body);
          return;
        } 
        if (typeof body !== 'string') {
          resolve(<T>body);
          return;
        }
        const content: GithubAPIGetResult = JSON.parse(<string>body);
        if (Array.isArray(content)) { // from keys return
          resolve(<T>content);
        } else {
          try {
            content.content = Buffer.from(<string>content.content, <"base64">content.encoding).toString('utf-8');
          } catch (e) {
            reject(new Error(`Do not support encoding "${content.encoding}"`));
            return;
          }
          resolve(<T>content);
        } 
      });
    })
  }

}
