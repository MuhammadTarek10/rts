export const ENVIRONMENT = {
  NODE_ENV: 'app.env',

  DATABASE: {
    KEY: 'database',
    URL: 'database.url',
  },

  JWT: {
    ACCESS_SECRET: 'jwt.access_secret',
    ACCESS_EXPIRATION: 'jwt.access_expiration',
    REFRESH_SECRET: 'jwt.refresh_secret',
    REFRESH_EXPIRATION: 'jwt.refresh_expiration',
  },
};

export const AUTH_STRATEGIES = {
  GOOGLE: 'google',
  GITHUB: 'github',
  FACEBOOK: 'facebook',
  LOCAL: 'local',
};
