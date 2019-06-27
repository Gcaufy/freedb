import { expect } from 'chai';
import request from 'request';
import KV from '../../src/kv';

const kv = new KV({
  host: 'git@github.com:Gcaufy-Test/test-database.git',
  token: '2bc3e8dc021e726417d77b7cb48c293cc2d820c0',
  debug: true,
  db: 'test',
});

const kvEncrypt = new KV({
  host: 'git@github.com:Gcaufy-Test/test-database.git',
  token: '2bc3e8dc021e726417d77b7cb48c293cc2d820c0',
  debug: true,
  db: 'test-encrypt',
  cipher: 'hello world'
});


const existKey = 'key-exist';
const nonExistKey = 'key-none-exist';

describe('KV', function() {

  it('keys', function (done) {
    kv.keys().then((res: any) => {
      var exist: boolean = false;
      res.forEach((item: any) => {
        if (item.name === existKey) {
          exist = true;
        }
      });
      expect(exist).equal(true);
      done();
    });
  });

  it('exist: exist', function (done) {
    kv.exist(existKey).then((res: boolean) => {
      expect(res).equal(true);
      done();
    });
  });

  it('exist: non-exist', function (done) {
    kv.exist(nonExistKey).then((res: boolean) => {
      expect(res).equal(false);
      done();
    });
  });
  it('set exist', function (done) {
    const value = 'value' + (Math.random() * 100000000 >> 0).toString()
    kv.set(existKey, value).then(res => {
      request.get(<string>res.raw_url + '?t=' + Math.random(), function (err: any, res: any, body: any) {
        expect(body).equal(value);
        done();
      })
    });
  });
  it('set: non-exist', function (done) {
    const value = 'value' + (Math.random() * 100000000 >> 0).toString()
    kv.set(nonExistKey, value).then(res => {
      request.get(<string>res.raw_url, function (err: any, res: any, body: any) {
        expect(body).equal(value);
        kv.delete(nonExistKey).then(() => {
          done();
        });
      })
    });
  });

  it('append: exist', function (done) {
    const value = 'value' + (Math.random() * 100000000 >> 0).toString()
    kv.append(existKey, value).then(res => {
      request.get(<string>res.raw_url, function (err: any, res: any, body: any) {
        expect(body).is.not.equal(value);
        done();
      })
    });
  });

  it('append: non-exist', function (done) {
    const value = 'value' + (Math.random() * 100000000 >> 0).toString()
    kv.append(nonExistKey, value).then(res => {
      request.get(<string>res.raw_url, function (err: any, res: any, body: any) {
        expect(body).equal(value);
        kv.delete(nonExistKey).then(() => { // revert it
          done();
        });
      })
    });
  });

  it('delete: exist', function(done) {
    const value = 'value' + (Math.random() * 100000000 >> 0).toString()
    kv.delete(existKey).then((res: any) => {
      kv.set(existKey, value).then(() => {
        done()
      }); // revert it
    });
  });

  it('delete: non-exist', function(done) {
    kv.delete(nonExistKey).then((res: any) => {
      expect(res.commit).equal('');
      done();
    });
  });

  it('use', function (done) {
    kv.use('other-db');
    kv.set('mykey', 'myvalue').then((res: any) => {
      request.get(<string>res.raw_url, function (err: any, res: any, body: any) {
        expect(body).equal('myvalue');
        kv.delete('mykey').then(() => {
          kv.use('some-other-non-exist-db');
          kv.keys().then((res: any) => {
            expect(res.length).equal(0);
            done();
          });
        });
      });
    })
  });

  it('encrypt kv', function (done) {
    kvEncrypt.set('encryptkey', 'encrypt-value').then((res: any) => {
      console.log(res);
      done()
    });
  })
});
