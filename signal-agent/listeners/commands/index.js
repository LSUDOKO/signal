import { signalCommandCallback } from './signal.js';
import { translateCommandCallback } from './translate.js';
import { catchupCommandCallback } from './catchup.js';
import { focusCommandCallback } from './focus.js';
import { digestCommandCallback } from './digest.js';

export const register = (app) => {
  app.command('/signal', signalCommandCallback);
  app.command('/translate', translateCommandCallback);
  app.command('/catchup', catchupCommandCallback);
  app.command('/focus', focusCommandCallback);
  app.command('/digest', digestCommandCallback);
};
