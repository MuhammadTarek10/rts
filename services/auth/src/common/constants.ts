export const ENVIRONMENT = {
  NODE_ENV: 'app.env',
  PORT: 'app.port',
  FRONTEND_URL: 'app.frontend_url',

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

  SWAGGER: {
    USERNAME: 'swagger.user',
    PASSWORD: 'swagger.password',
  },

  BROKER: {
    KEY: 'broker',
    URI: 'broker.RABBITMQ_URI',
    QUEUES: {
      AUTH_EVENTS: 'auth.events',
    },
    EVENTS: {
      USER_REGISTERED: 'UserRegistered',
    },
  },
};

export const AUTH_STRATEGIES = {
  GOOGLE: 'google',
  GITHUB: 'github',
  FACEBOOK: 'facebook',
  LOCAL: 'local',
};

export const USER_STATUS = {
  PENDING: 'pending',
  ACTIVE: 'active',
  INACTIVE: 'inactive',
  SUSPENDED: 'suspended',
};

export const USER_ROLES = {
  USER: 'user',
  ADMIN: 'admin',
};

export const LOCKDOWN_THRESHOLD = 5; // failed attempts before lockdown
export const LOCKDOWN_DURATION = 15 * 60 * 1000; // 15 minutes in milliseconds
