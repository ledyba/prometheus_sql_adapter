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
pub enum Repo {
  Sqlite(SqliteRepo)
}

#[derive(Clone)]
pub struct SqliteRepo {
  pool: sqlx::SqlitePool,
}

impl SqliteRepo {
  pub async fn init(&mut self) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    let _create_result = sqlx::query(r"
create table timeseries(
  id integer primary key autoincrement
);
create table labels(
  id integer primary key autoincrement,
  timeseries_id integer,
  name text,
  value text
);
create table samples(
  id integer primary key autoincrement,
  timeseries_id integer,
  timestamp integer,
  value real
);
create index labels_timestamp_index on labels(timestamp);
create index samples_timestamp_index on samples(timestamp);
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

impl Repo {
  pub async fn init(&mut self) -> sqlx::Result<()> {
    match self {
      Repo::Sqlite(repo) => repo.init().await
    }
  }
  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    match self {
      Repo::Sqlite(repo) => repo.write(req).await
    }
  }
  pub async fn read(&mut self, query: Query) -> sqlx::Result<QueryResult> {
    match self {
      Repo::Sqlite(repo) => repo.read(query).await
    }
  }
}


pub async fn open(url: &str) -> std::result::Result<Repo, Box<dyn std::error::Error>> {
  match url {
    url if url.starts_with("sqlite:") => {
      let pool = sqlx::sqlite::SqlitePool::builder().build(url).await.map_err(|err| err)?;
      Ok(Repo::Sqlite(SqliteRepo{
        pool,
      }))
    }
    url => Err(string_error::new_err(format!("Unsupportd DB: {}", url).as_str())),
  }
}
