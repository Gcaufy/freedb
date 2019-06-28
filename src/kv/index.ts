import crypto from 'crypto';
import { Host, parseHost } from '../helper';
import { Querier, QuerierOption, KeyRecord, Committer } from './Querier';


interface EncryptOption {
  secret: string;
  encode (str: string): string;
  decode (str: string): string;
}

type CipherOption = string | {
  secret: string;
  encode? (str: string): string;
  decode? (str: string): string;
}


interface DataBaseOption {
  // host support both https link and ssh link
  host: string;
  // generate a token from: github.com -> Settings -> Developer settings -> Personal access tokens -> Generate new token
  // which link is: https://github.com/settings/tokens
  token: string;
  // default db is 'default'
  db?: string;
  // default branch is 'master'
  branch?: string;
  // encrypt the key and the value
  cipher?: CipherOption;
  // show debug log, default to false
  debug?: boolean;
  // git committer
  committer?: Committer
}

interface Cache {
  [key: string]: KeyRecord 
}

/**
 * KV class. A basic key-value database based on github.com.
 *
 * @example
 * const kv = new KV({ host: 'git@github.com/Gcaufy/test-db.git', token: 'xxxxxxxx' })
 * await kv.set('mykey', 'myvalue')
 * let value = await kv.get('mykey');
 * console.log(value);
 * await kv.delete('mykey');
 */
export default class KV {

  public readonly option: DataBaseOption

  private readonly host: Host

  private querier: Querier

  private cache: Cache  = {}
  private cacheList: KeyRecord[] = []
  private encrypt: boolean = false
  private cipher: EncryptOption = { 
    secret: '', 
    encode: ((str: string) => str), 
    decode: ((str: string) => str)
  }

  /**
   * @private
   */
  constructor (option: DataBaseOption) {
    if (!option.host) {
      throw new Error('Host is unset');
    }
    if (!option.token) {
      throw new Error('Token is unset. Please check here: https://developer.github.com/v3/oauth_authorizations/#create-a-new-authorization');
    }

    this.host = parseHost(option.host);
    this.option = {
      db: 'default',
      branch: 'master',
      debug: false,
      committer: <Committer>{
        name: 'freedb',
        email: 'freedb@unknow.email',
      },
      ...option
    };

    this.initCipher();

    const QuerierConstructor: { new(option: QuerierOption): Querier } = require('./' + this.host.provider).default;
    const querierOption: QuerierOption = {
      user: this.host.user,
      repo: this.host.repo,
      token: this.option.token,
      db: <string>this.option.db,
      branch: <string>this.option.branch,
      debug: <boolean>this.option.debug,
      committer: <Committer>this.option.committer,
    };
    this.querier = new QuerierConstructor(querierOption);
  }

  private initCipher (): void {
    const cipher = this.option.cipher;
    if (cipher) {
      if (typeof cipher === 'string') {
        this.cipher = <EncryptOption>{
          secret: cipher
        };
      } else if (typeof cipher === 'object') {
        this.cipher = <EncryptOption>{
          secret: cipher.secret,
        };
      }
      if (!this.cipher.encode || !this.cipher.decode) {
        this.cipher.encode = function (str: string): string {
          const cipher = crypto.createCipher('aes192', this.secret);
          cipher.update(str, 'utf8', 'hex');
          return cipher.final('hex');
        };
        this.cipher.decode = function (str): string {
          const cipher = crypto.createDecipher('aes192', this.secret);
          cipher.update(str, 'hex', 'utf8');
          return cipher.final('utf-8');
        };  
      }
      this.encrypt = true;
    } else {
      this.encrypt = false;
    }
  }

  /**
   * Change current db
   * @param {string} db
   * @example
   * kv.use('test-db');
   * await kv.set('mynewkey', '123')
   */
  use (db: string): void {
    // clear the cache
    this.cache = {};
    this.cacheList = [];
    this.querier.db(db);
  }

  /**
   * Set value for a key. If the key do not exist, then create a new key, otherwise update the key value
   * @param {string} key
   * @param {string} value
   * @return {Promise<KeyRecord>}
   * @example
   * await kv.set('mykey', 'append-value');
   */
  set (key: string, value: string): Promise<KeyRecord> {
    const setkey = this.encrypt ? this.cipher.encode(key) : key;
    const setval = this.encrypt ? this.cipher.encode(value) : value;
    return this.querier.set(setkey, setval).then((res: KeyRecord) => {
      if (this.encrypt) {
        res.content = this.cipher!.decode(res.content || '');
        res.name = this.cipher!.decode(res.name || '');
      }
      this.cache[key] = res;
      return res;
    });
  }

  /**
   * Append value to a key, if the key do not exist, then create a new key. Otherwise append value to it.
   * @param {string} key
   * @param {string} value
   * @return {Promise<KeyRecord>}
   * @example
   * await kv.append('mykey', 'append-value');
   */
  append (key: string, value: string): Promise<KeyRecord> {
    return this.get(key).then((res: KeyRecord) => {
      value = res.content + value;
      return this.set(key, value);
    });
  }

  /**
   * List all keys in database
   * @returns {Promise<KeyRecord[]>}
   * @example
   * let keys = await kv.keys();
   * keys.forEach(key => console.log(key.name));
   */
  keys (): Promise<KeyRecord[]> {
    return new Promise((resolve, reject) => {
      if (this.cacheList.length) {
        resolve(this.cacheList);
      } else {
        return this.querier.keys().then((res: KeyRecord[]) => {
          this.cacheList = this.encrypt 
            ? res.map((v) => {
                v.content = this.cipher.decode(v.content || '');
                v.name = this.cipher.decode(v.name || '');
                return v;
              })
            : res;
          resolve(this.cacheList);
        });
      }
    });
  }

  /**
   * Check is a key exist or not
   * @param {string} key
   * @returns {Promise<boolean>}
   * @example
   * var exist = await kv.exist('mykey');
   */
  exist (key: string): Promise<boolean>  {
    return this.get(key).then((res: KeyRecord) => {
      return res.size !== -1;
    });
  }

  /**
   * Get value of my key
   * @param {string} key
   * @returns {Promise<KeyRecord>}
   * @example
   * var key = await kv.get('mykey');
   * console.log(key.content);
   */
  get (key: string): Promise<KeyRecord> {
    return new Promise((resolve, reject) => {
      const cache = this.cache[key];
      if (cache) {
        resolve(cache); 
      } else {
        const getkey = this.encrypt ? this.cipher.encode(key) : key;
        this.querier.get(getkey).then((res: KeyRecord) => {
          if (this.encrypt) {
            res.content = this.cipher!.decode(res.content || '');
            res.name = this.cipher!.decode(res.name || '');
          }
          this.cache[key] = res;
          resolve(res);
        });
      }
    });
  }

  /**
   * Delete a key
   * @param {string}key
   * @returns {Promise<{ commit: string }>}
   * @example
   * await kv.delete('mykey')
   */
  delete (key: string): Promise<KeyRecord> {
    const getkey = this.encrypt ? this.cipher.encode(key) : key;
    return this.querier.delete(key).then((res: KeyRecord) => {
      delete this.cache[key];
      return res;
    });
  }
}
