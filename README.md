# freedb 

A lightweight solution to use a cloud Key-Value database based on github.com.


## Install

```
npm install freedb --save
```

## Usage

```
import Free from 'freedb';

const kv = Free.KV({
  // This is my public test account and token, only used for test and CI.
  // If you want to have your own database,
  // Then you need to create a repository and generate a token.
  host: 'git@github.com:Gcaufy-Test/test-database.git',
  token: 'your_github_repo_token'
});

kv.set('mykey', 'myvalue').then(res => {
  console.log(res.raw_url);
});
```


## How to generate a token

  1. Create or use your own github.com account and login.
  2. Go to: Settings -> Developer settings -> Personal access tokens -> Generate new token
  3. Select scopes: "repo" to make sure you grant access.


## API:


* Create a KV instance

    new Free.KV(option: DataBaseOption):

    ```
    DataBaseOption {
        // host: github clone links, support both https/ssh links
        host: string;
        // token: OAuth token, make sure you have read/write access for the repo
        token: string;
        // db: Basiclly it's a directory, default value is "default"
        db?: string;
        // branch: git branch, default value is "master"
        branch?: string;
        // cipher: if is a string, then treat as a secret key for aes192 for both key and value. or you can customize a encryt and decrypt function
        cipher?: CipherOption | string;
        // debug: show action log or not.
        debug?: boolean;
    }

    CipherOption {
      // secret key for encrypt
      secret: string;
      // customize encrypt algorithm, default value is ase192 encrypt algorithm
      encode: (str: string): string;
      // customize decrypt algorithm, default value is ase192 decrypt algorithm
      decode: (str: string): string;
    }
    ```

* KV instance methods

    1. use(db: string): void

    Switch database. can be a non-exist database.

    2. keys(): Promise<KeyRecord[]>

    List all keys in current database;

    3. exist(key: string): Promise<boolean>

    Check a key exist or not in current database;

    4. get(key: string): Promise<KeyRecord>

    Get a key record in current database;

    5. set(key: string, value: string): Promise<KeyRecord>

    Set a value for a key. Will create a key if a key do not exist;

    6. append(key: string): Promise<KeyRecord>

    Append a value for a key. Will create a key if a key do not exist;


    ```
    KeyRecord {
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
    ```

## How to protect your data

There are two way to protect your data.

1. Encrypt you key and value in `CipherOption`
```
new GitDB.KV({
  host: 'git@github.com:Gcuafy-Test/test-database',
  token: 'mytoken',
  cipher: {
    secret: 'my secret key',
    // Default value is using ase192 encrypt algorithm
    encode (str) {
      return MyEncryptMethod(str);
    },
    // Default value is using ase192 decrypt algorithm
    decode (str) {
      return MyDecryptMethod(str);
    }
  }
})
```

2. Make the repository private.
  Simply and easy. Github support private repository
