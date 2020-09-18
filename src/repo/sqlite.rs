/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
use crate::proto::remote::{WriteRequest, Query, QueryResult};
use sqlx::sqlite::Sqlite;
use sqlx::pool::PoolConnection;

#[derive(Clone)]
pub struct Repo {
  pool: sqlx::SqlitePool,
}

impl Repo {
  pub fn new(pool: sqlx::SqlitePool) -> Repo {
    Repo{
      pool,
    }
  }

  pub async fn init(&mut self) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    let _create_result = sqlx::query(r"
create table if not exists timeseries(
  id integer primary key not null
);

create table if not exists labels(
  timeseries_id integer not null,
  name integer not null,
  value integer not null
);

create table if not exists literals(
  id integer primary key not null,
  value text unique not null
);

create table if not exists samples(
  timeseries_id integer not null,
  timestamp integer not null,
  value real not null
);

-- labels
create index if not exists labels_timeseries_index on labels(timeseries_id);

-- samples
create index if not exists samples_timestamp_index on samples(timestamp);
create index if not exists samples_timeseries_index on samples(timeseries_id);

-- literals
create index if not exists literals_value_index on literals(value);
").execute(&mut conn).await?;
    Ok(())
  }

  async fn try_write(&mut self, mut conn: PoolConnection<Sqlite>, id: i64, req: &WriteRequest) -> sqlx::Result<()> {
    for ts in req.timeseries.iter() {
      for sample in ts.samples.iter() {
        sqlx::query::<Sqlite>(r"insert into samples (timeseries_id, timestamp, value) values (?, ?, ?)")
          .bind(id)
          .bind(sample.timestamp)
          .bind(sample.value)
          .execute(&mut conn)
          .await?;
      }
      for label in ts.labels.iter() {
        sqlx::query::<Sqlite>(r"insert into labels (timeseries_id, name, value) values (?, (select id from literals where value = ?), (select id from literals where value = ?))")
          .bind(id)
          .bind(label.name.as_str())
          .bind(label.value.as_str())
          .execute(&mut conn)
          .await?;
      }
    }
    Ok(())
  }

  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    {
      for ts in req.timeseries.iter() {
        for label in ts.labels.iter() {
          sqlx::query::<Sqlite>(r"insert or ignore into literals (id, value) values (?, ?), (?, ?)")
            .bind(rand::random::<i64>())
            .bind(label.name.as_str())
            .bind(rand::random::<i64>())
            .bind(label.value.as_str())
            .execute(&mut conn)
            .await?;
        }
      }
    }
    let id = {
      let mut id:i64 = 0;
      for _ in 0..10 {
        id = rand::random::<i64>();
        let result = sqlx::query("insert into timeseries (id) values (?)").bind(id).execute(&mut conn).await;
        if result.is_ok() {
          break;
        }
      }
      id
    };
    let result = self.try_write(conn, id, &req).await;
    if result.is_err() {
      let mut conn = self.pool.acquire().await?;
      sqlx::query::<Sqlite>(r##"delete from timeseries where id = ?"##)
        .bind(id)
        .execute(&mut conn)
        .await?;
    }
    Ok(())
  }

  pub async fn read(&mut self, query: &Query) -> sqlx::Result<QueryResult> {
    // FIXME
    Err(sqlx::Error::RowNotFound)
  }
}

