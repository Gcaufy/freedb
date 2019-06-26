export interface Logger {
  log (cateogry: string, msg: string): void
}
export function createLogger (option: { debug: boolean }): Logger {
  var log = (category: string, msg: string): void => {}
  if (option.debug) {
    log = (category: string, msg: string): void => {
      console.log(`[${category}]: ${msg}`);
    }
  }
  return {
    log
  };
}
