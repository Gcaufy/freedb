# freedb 

A lightweight solution to use a cloud Key-Value database based on github.com.

![freedb](https://user-images.githubusercontent.com/2182004/60488728-50dae280-9cd5-11e9-933b-b6798f87af95.png)

## Install

```
go get -u github.com/Gcaufy/freedb
```

## Build from source

See [BUILD.md](BUILD.md).

## Usage

```
 $ freedb -?
Usage:
  freedb [flags]

Flags:
  -b, --branch string     Config using branch. (default "master")
  -d, --database string   Config using database. (default "default")
  -e, --execute string    Execute command and quit.
  -?, --help              Display the help
  -h, --host string       Connect to host, which is a https/ssh git clone link.
  -s, --short-output      Only output the value
  -t, --token string      Access token for the git repository.
```


## How to generate a token

  1. Create or use your own github.com account and login.
  2. Go to: Settings -> Developer settings -> Personal access tokens -> Generate new token
  3. Select scopes: "repo" to make sure you grant access.


## API:
 Working on it

## How to protect your data

1. Make the repository private.
  Simply and easy. Github support private repository
