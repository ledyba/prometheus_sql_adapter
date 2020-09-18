/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/

use crate::proto::remote::{Query, QueryResult, WriteRequest};

mod sqlite;
mod mysql;

#[derive(Clone)]
pub enum Repo {
  Sqlite(sqlite::Repo),
  MySql(mysql::Repo),
}

impl Repo {
  pub async fn init(&mut self) -> sqlx::Result<()> {
    match self {
      Repo::Sqlite(repo) => repo.init().await,
      Repo::MySql(repo) => repo.init().await,
    }
  }
  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    match self {
      Repo::Sqlite(repo) => repo.write(req).await,
      Repo::MySql(repo) => repo.write(req).await,
    }
  }
  pub async fn read(&mut self, query: &Query) -> sqlx::Result<QueryResult> {
    match self {
      Repo::Sqlite(repo) => repo.read(query).await,
      Repo::MySql(repo) => repo.read(query).await,
    }
  }
}


pub async fn open(url: &str) -> std::result::Result<Repo, Box<dyn std::error::Error>> {
  match url {
    url if url.starts_with("sqlite:") => {
      let pool = sqlx::sqlite::SqlitePool::connect(url)
        .await.map_err(|err| err)?;
      Ok(Repo::Sqlite(sqlite::Repo::new(pool)))
    }
    url if url.starts_with("mysql:") => {
      let pool = sqlx::mysql::MySqlPool::connect(url)
        .await.map_err(|err| err)?;
      Ok(Repo::MySql(mysql::Repo::new(pool)))
    }
    url => Err(string_error::new_err(format!("Unsupportd DB: {}", url).as_str())),
  }
}