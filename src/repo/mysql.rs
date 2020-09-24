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
  `id` bigint unsigned auto_increment not null,
  primary key (`id`)
) ENGINE=InnoDB;").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists `labels`(
  `timeseries_id` bigint unsigned not null,
  `name` int not null,
  `value` int not null,
  index (`timeseries_id`)
) ENGINE=InnoDB;").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists `literals`(
  `id` bigint unsigned auto_increment not null,
  `value` varchar(256) unique not null,
  primary key (`id`),
  index (`value`)
) ENGINE=InnoDB;").execute(&mut conn).await?;

    sqlx::query(r"
create table if not exists samples(
  `timeseries_id` bigint unsigned not null,
  `timestamp` bigint unsigned not null,
  `value` double not null,
  index (`timeseries_id`),
  index (`timestamp`)
) ENGINE=InnoDB;").execute(&mut conn).await?;

    Ok(())
  }

  pub async fn write(&mut self, req: WriteRequest) -> sqlx::Result<()> {
    {
      let mut conn = self.pool.acquire().await?;
      let mut places: Vec<&str> = vec![];
      let mut words: Vec<&String> = vec![];
      let mut sql = r"insert ignore into `literals` (`value`) values ".to_string();
      for ts in req.timeseries.iter() {
        for label in ts.labels.iter() {
          places.push("(?),(?)");
          words.push(&label.name);
          words.push(&label.value);
        }
      }
      sql += places.join(", ").as_str();
      let query = sqlx::query::<MySql>(sql.as_str());
      words.into_iter().fold(query, |query, word| query.bind(word))
        .execute(&mut conn)
      .await?;
    }
    //let mut conn = self.pool.begin().await?;
    let mut conn = self.pool.acquire().await?;
    let mut sample_sql = r"insert into samples (timeseries_id, timestamp, value) values ".to_string();
    let mut sample_places:Vec<&str> = vec![];
    let mut label_sql = r"insert into labels (timeseries_id, name, value) values ".to_string();
    let mut label_places:Vec<&str> = vec![];
    for ts in req.timeseries.iter() {
      for _ in ts.samples.iter() {
        sample_places.push("(?, ?, ?)");
      }
      for _ in ts.labels.iter() {
        label_places.push("(?, (select id from literals where value = ?), (select id from literals where value = ?))");
      }
    }
    sample_sql += sample_places.join(", ").as_str();
    label_sql += label_places.join(", ").as_str();
    let mut sample_query = sqlx::query::<MySql>(sample_sql.as_str());
    let mut label_query = sqlx::query::<MySql>(label_sql.as_str());
    for ts in req.timeseries.iter() {
      let id: u64 = {
        sqlx::query::<MySql>(r"insert into timeseries () values ()")
          .execute(&mut conn).await?;
        sqlx::query_as::<MySql, (u64,)>("select last_insert_id()")
          .fetch_one(&mut conn).await?.0
      };
      for sample in ts.samples.iter() {
        sample_query = sample_query
          .bind(id)
          .bind(sample.timestamp)
          .bind(sample.value);
      }
      for label in ts.labels.iter() {
        label_query = label_query
          .bind(id)
          .bind(label.name.as_str())
          .bind(label.value.as_str());
      }
    }
    sample_query.execute(&mut conn).await?;
    label_query.execute(&mut conn).await?;
    //conn.commit().await
    Ok(())
  }

  pub async fn read(&mut self, query: &Query) -> sqlx::Result<QueryResult> {
    // FIXME
    Err(sqlx::Error::RowNotFound)
  }
}
