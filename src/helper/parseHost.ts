export interface Host {
  provider: string;
  user: string;
  repo: string;
}
export function parseHost (host: string): Host {
  const regList: Array<RegExp> = [
    /git@([\w\.]+):([\w_-]+)\/([\w\._-]+)\.git/,
    /https:\/\/([\w\.]+)\/([\w_-]+)\/([\w\._-]+)\.git/,
  ];

  for (const reg of regList) {
    const matchs = host.match(reg);
    if (matchs) {
      return {
        provider: matchs[1],
        user: matchs[2],
        repo: matchs[3],
      };
    }
  }

  throw new Error(`Can not reconigize host: ${host}`);
}
