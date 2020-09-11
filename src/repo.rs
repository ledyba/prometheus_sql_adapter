/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/

use crate::proto::remote::{WriteRequest, Query, QueryResult};
use sqlx::sqlite::SqliteCursor;
use sqlx::prelude::*;

#[derive(Clone)]
pub enum Repo {
  Sqlite(SqliteRepo)
}

#[derive(Clone)]
struct SqliteRepo {
  pool: sqlx::SqlitePool,
}

impl SqliteRepo {
  pub async fn init(&mut self) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    sqlx::query(r"
create table timeseries(
  id integer primary key autoincrement
);
create table label(
  timeseries_id integer,
  name text,
  value text
);
create table sample(
  id integer primary key autoincrement,
  timeseries_id integer,
  timestamp: integer,
  value real
);
").execute(&mut conn).await?;
    Ok(())
  }
  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    let mut tx = self.pool.begin().await?;
    let _ = tx.execute("insert into timeseries () values ()").await?;
    let mut cur: SqliteCursor = tx.fetch("select id from timeseries where rowid = last_insert_rowid()");
    let id: i64 = cur.next().await?.expect("Failed to insert timeseries").get(0);
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
