/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/

use cascara::Cache;
use tokio::sync::RwLock;
use crate::repo::Repo;

pub struct Context {
  pub cache: RwLock<Cache<String, String>>,
  pub db: Repo,
}