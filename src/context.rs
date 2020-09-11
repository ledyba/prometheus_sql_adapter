/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/

use tokio::sync::RwLock;
use std::cell::RefCell;
use std::sync::Arc;

pub struct Context {
  pub cache: RwLock<cascara::Cache<String, String>>,
  pub db: crate::repo::Repo,
}