/**
 * Logger utility that only logs in development mode
 * In production, all log methods are no-ops to prevent console pollution
 */
const isDev = import.meta.env.DEV;

export const logger = {
  log: isDev ? console.log.bind(console) : () => {},
  error: isDev ? console.error.bind(console) : () => {},
  warn: isDev ? console.warn.bind(console) : () => {},
  info: isDev ? console.info.bind(console) : () => {},
  debug: isDev ? console.debug.bind(console) : () => {},
};

// For backward compatibility, also export as default
export default logger;
