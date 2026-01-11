/**
 * Database Connection
 * 
 * PostgreSQL connection with TypeORM support.
 */

import { DataSource, EntityManager } from 'typeorm';
import { config } from '../config/config';

let dataSource: DataSource | null = null;

export function getDataSource(): DataSource {
  if (dataSource) {
    return dataSource;
  }

  dataSource = new DataSource({
    type: 'postgres',
    host: config.database.host,
    port: config.database.port,
    username: config.database.username,
    password: config.database.password,
    database: config.database.name,
    ssl: config.database.ssl_mode === 'require' ? { rejectUnauthorized: false } : false,
    entities: [__dirname + '/../models/*.ts'],
    migrations: [__dirname + '/../migrations/*.ts'],
    synchronize: config.app.environment === 'development',
    logging: config.app.environment === 'development',
    extra: {
      max: config.database.max_open_conns,
      idleTimeoutMillis: config.database.conn_max_lifetime * 1000,
    },
  });

  return dataSource;
}

export async function initializeDatabase(): Promise<void> {
  const ds = getDataSource();
  if (!ds.isInitialized) {
    await ds.initialize();
    console.log('Database connected');
  }
}

export async function closeDatabase(): Promise<void> {
  if (dataSource && dataSource.isInitialized) {
    await dataSource.destroy();
    console.log('Database connection closed');
  }
}

export function getManager(): EntityManager {
  return getDataSource().manager;
}
