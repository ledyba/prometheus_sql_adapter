/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
use crate::proto::remote::{WriteRequest, Query, QueryResult};
use sqlx::prelude::*;
use sqlx::{Transaction, SqliteConnection, Sqlite};
use sqlx::pool::PoolConnection;

#[derive(Clone)]
pub struct Repo {
  pool: sqlx::SqlitePool,
}

impl Repo {
  pub fn new(pool: sqlx::SqlitePool) -> Repo {
    Repo{
      pool
    }
  }

  pub async fn init(&mut self) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    let _create_result = sqlx::query(r"
create table if not exists timeseries(
  id integer primary key autoincrement
);
create table if not exists labels(
  id integer primary key autoincrement,
  timeseries_id integer,
  name text,
  value text
);
create table if not exists samples(
  id integer primary key autoincrement,
  timeseries_id integer,
  timestamp integer,
  value real
);
create index if not exists labels_timeseries_index on labels(timeseries_id);
create index if not exists samples_timestamp_index on samples(timestamp);
create index if not exists samples_timeseries_id_index on samples(timeseries_id);
").execute(&mut conn).await?;
    Ok(())
  }
  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    let mut tx: Transaction<PoolConnection<SqliteConnection>> = self.pool.begin().await?;
    for ts in req.timeseries.iter() {
      let _ = sqlx::query::<Sqlite>("insert into timeseries default values").execute(&mut tx).await?;
      let id: (i64,) = SqliteQueryAs::fetch_one(sqlx::query_as("select id from timeseries where rowid = last_insert_rowid()"), &mut tx).await?;
      for sample in ts.samples.iter() {
        sqlx::query::<Sqlite>(r"insert into samples (timeseries_id, timestamp, value) values (?, ?, ?)")
          .bind(id.0)
          .bind(sample.timestamp)
          .bind(sample.value)
          .execute(&mut tx)
          .await?;
      }
      for label in ts.labels.iter() {
        sqlx::query::<Sqlite>(r"insert into labels (timeseries_id, name, value) values (?, ?, ?)")
          .bind(id.0)
          .bind(label.name.as_str())
          .bind(label.value.as_str())
          .execute(&mut tx)
          .await?;
      }
    }
    tx.commit().await?;
    Ok(())
  }
  pub async fn read(&mut self, query: Query) -> sqlx::Result<QueryResult> {
    // FIXME
    Err(sqlx::Error::RowNotFound)
  }
}

