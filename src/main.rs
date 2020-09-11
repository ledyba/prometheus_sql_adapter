/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
use std::str::FromStr;
use std::sync::{Arc};
use std::process::exit;

#[macro_use]
extern crate log;
use env_logger::Env;

use clap::{App, Arg, SubCommand, ArgMatches};

use warp::Filter;
use warp::hyper::body::Bytes;

mod context;
mod handlers;
mod proto;
mod repo;

fn web(m: &ArgMatches) -> Result<(), Box<dyn std::error::Error>> {
  let sock = if let Some(listen) = m.value_of("listen") {
    std::net::SocketAddr::from_str(listen)?
  } else {
    return Err("listen is not set.".into())
  };
  let db_uri = if let Some(db_uri) = m.value_of("db") {
    db_uri
  } else {
    return Err("listen is not set.".into())
  };

  let init = m.values_of("init").is_some();

  let mut rt = tokio::runtime::Builder::new()
    .core_threads(32)
    .threaded_scheduler()
    .enable_all()
    .build()
    .unwrap();

  rt.block_on(async {
    let mut db = repo::open(db_uri).await?;
    if init {
      match db.init().await {
        Ok(()) => info!("Database initialized!"),
        Err(err) => warn!("Failed to init DB: {:?}", err),
      };
    }

    let conf = Arc::new(context::Context {
      cache: tokio::sync::RwLock::new(cascara::Cache::with_window_size(100, 20)),
      db,
    });

    let write_conf = conf.clone();
    let write_handler = move |body: Bytes| handlers::write(write_conf.clone(), body);
    let writer = warp::post()
      .and(warp::path("write"))
      .and(warp::body::content_length_limit(1024 * 1024 * 16))
      .and(warp::body::bytes())
      .and_then(write_handler);
    let read_conf = conf.clone();
    let read_handler = move |body: Bytes| handlers::read(read_conf.clone(), body);
    let reader = warp::post()
      .and(warp::path("read"))
      .and(warp::body::content_length_limit(1024 * 1024 * 16))
      .and(warp::body::bytes())
      .and_then(read_handler);
    let index = warp::path::end().and_then(handlers::not_found);
    let router = index
      .or(writer)
      .or(reader)
      .or(warp::any().and_then(handlers::not_found));
    warp::serve(router)
      .run(sock)
      .await;
    Ok(())
  })
}

fn main() {
  env_logger::from_env(Env::default().default_filter_or("info")).init();

  let app = App::new("prometheus_sql_adapter")
    .version("0.1.0")
    .author("Kaede Fujisaki <psi@7io.org>")
    .about("Prometheus SQL Remote Storage Adapter for Generic RDBMS")
    .subcommand(SubCommand::with_name("web")
      .arg(Arg::with_name("listen")
        .long("listen")
        .takes_value(true)
        .allow_hyphen_values(true)
        .default_value("0.0.0.0:8080")
        .required(false))
      .arg(Arg::with_name("db")
        .long("db")
        .takes_value(true)
        .allow_hyphen_values(true)
        .default_value("sqlite:")
        .required(false))
      .arg(Arg::with_name("init")
        .long("init")
        .takes_value(false)
        .required(false)));
  let m = app.get_matches();
  match m.subcommand_name() {
    Some("web") => {
      if let Err(err) = web(m.subcommand_matches("web").unwrap()) {
        error!("Failed to start web: {:?}\n", err);
        exit(-1);
      }
    }
    None | Some(_) => {
      error!("{}\n", m.usage());
      exit(-1);
    }
  }
}