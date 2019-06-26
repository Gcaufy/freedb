# GitDB

A lightweight solution to use a cloud Key-Value database based on github.com.


## Install

```
npm install git-db --save
```

## Usage

```
import GitDB from 'git-db';

const kv = GitDB.KV({
  // This is my public test account and token, only used for test and CI.
  // If you want to have your own database,
  // Then you need to create a repository and generate a token.
  host: 'git@github.com:Gcaufy-Test/test-database.git',
  token: '69ccdeabf7517f8c44b3ce37cae7480b4ae25fe9'
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

* new GitDB.KV(option: DataBaseOption):

    Create a kv instance
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
    ```

>>>>>>> init code

