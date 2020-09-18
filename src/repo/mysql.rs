/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
use crate::proto::remote::{WriteRequest, Query, QueryResult};
use sqlx::mysql::{MySql, MySqlPool};

#[derive(Clone)]
pub struct Repo {
  pool: MySqlPool,
}

impl Repo {
  pub fn new(pool: MySqlPool) -> Repo {
    Repo{
      pool,
    }
  }

  pub async fn init(&mut self) -> sqlx::Result<()> {
    let mut conn = self.pool.acquire().await?;
    sqlx::query(r"
create table if not exists `timeseries`(
  `id` int unsigned auto_increment not null,
  primary key (`id`)
);").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists `labels`(
  `timeseries_id` int unsigned not null,
  `name` int not null,
  `value` int not null,
  index (`timeseries_id`)
);").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists `literals`(
  `id` int unsigned auto_increment not null,
  `value` varchar(512) not null,
  primary key (`id`),
  index (`value`)
);").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists samples(
  `timeseries_id` int unsigned not null,
  `timestamp` bigint unsigned not null,
  `value` double not null,
  index (`timeseries_id`),
  index (`timestamp`)
);").execute(&mut conn).await?;

    Ok(())
  }

  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    {
      let mut conn = self.pool.acquire().await?;
      for ts in req.timeseries.iter() {
        for label in ts.labels.iter() {
          sqlx::query::<MySql>(r"insert ignore into `literals` (`value`) values (?), (?)")
            .bind(label.name.as_str())
            .bind(label.value.as_str())
            .execute(&mut conn)
            .await?;
        }
      }
    }
    let mut conn = self.pool.begin().await?;
    let id: u64 = {
      sqlx::query_as::<MySql, (u64,)>("insert into timeseries default values; select last_insert_id(`id`)")
        .fetch_one(&mut conn).await?.0
    };
    for ts in req.timeseries.iter() {
      for sample in ts.samples.iter() {
        sqlx::query::<MySql>(r"insert into samples (timeseries_id, timestamp, value) values (?, ?, ?)")
          .bind(id)
          .bind(sample.timestamp)
          .bind(sample.value)
          .execute(&mut conn)
          .await?;
      }
      for label in ts.labels.iter() {
        sqlx::query::<MySql>(r"insert into labels (timeseries_id, name, value) values (?, (select id from literals where value = ?), (select id from literals where value = ?))")
          .bind(id)
          .bind(label.name.as_str())
          .bind(label.value.as_str())
          .execute(&mut conn)
          .await?;
      }
    }
    conn.commit().await
  }

  pub async fn read(&mut self, query: &Query) -> sqlx::Result<QueryResult> {
    // FIXME
    Err(sqlx::Error::RowNotFound)
  }
}

