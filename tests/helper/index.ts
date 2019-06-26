import { Host, parseHost } from '@/helper/parseHost';
import { expect } from 'chai';

describe('parseHost', function() {
  it('base', function() {

    const cases:Array<Host> = [
      {
        provider: 'github.com',
        user: 'Gcaufy',
        repo: 'test-repo'
      }
    ]
    for (const c of cases) {
      const git: string = `git@${ c.provider }:${ c.user }\/${ c.repo }.git`;
      const https: string = `https:\/\/${c.provider}/${c.user}\/${c.repo}.git`;

      const gitHost: Host = parseHost(git);
      expect(gitHost).to.deep.equal(c);

      const httpsHost: Host = parseHost(https);
      expect(httpsHost).to.deep.equal(c);

    }
  }); 

  it('error', () => {
    expect(() => { parseHost('abc') }).to.throw();
  });
});
