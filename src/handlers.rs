/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
use std::str::FromStr;
use std::sync::Arc;

extern crate warp;
use warp::hyper::body::{Bytes, Body};
use warp::reply::Reply;
use warp::reject;
use warp::http::uri;
use warp::http::StatusCode;
use warp::reply::Response;

extern crate snap;
use snap::raw::{Decoder, Encoder};

use crate::context;
use crate::proto;
use crate::proto::remote::{WriteRequest, ReadRequest, ReadResponse};
use protobuf::Message;

pub async fn not_found() -> Result<impl Reply, reject::Rejection> {
  Ok(warp::redirect(uri::Uri::from_str("/").unwrap()))
}

fn create_error_response(code: u16, err: impl std::error::Error) -> Response {
  warp::http::Response::builder()
    .status(code)
    .header("Content-Type", "text/plain;charset=UTF-8")
    .body(Body::from(err.to_string()))
    .unwrap()
}

pub async fn write(conf: Arc<context::Context>, body: Bytes) -> Result<impl Reply, reject::Rejection> {
  let mut decoder = snap::raw::Decoder::new();
  let decoding_result: Result<Vec<u8>,snap::Error> = decoder.decompress_vec(&body.to_vec());
  if decoding_result.is_err() {
    return Ok(create_error_response(400, decoding_result.unwrap_err()));
  }
  let data = decoding_result.unwrap();
  let proto_parse_result = protobuf::parse_from_bytes::<WriteRequest>(&data);
  if proto_parse_result.is_err() {
    return Ok(create_error_response(400, proto_parse_result.unwrap_err()));
  }
  let req = proto_parse_result.unwrap();
  for ts in req.timeseries.iter() {
  }
  Ok(warp::reply::html("OK").into_response())
}

pub async fn read(body: Bytes) -> Result<impl Reply, reject::Rejection> {
  // parse response
  let mut decoder = snap::raw::Decoder::new();
  let decoding_result: Result<Vec<u8>,snap::Error> = decoder.decompress_vec(&body.to_vec());
  if decoding_result.is_err() {
    return Ok(create_error_response(400, decoding_result.unwrap_err()));
  }
  let data = decoding_result.unwrap();
  let proto_parse_result = protobuf::parse_from_bytes::<ReadRequest>(&data);
  if proto_parse_result.is_err() {
    return Ok(create_error_response(400, proto_parse_result.unwrap_err()));
  }

  // create response
  let req = proto_parse_result.unwrap();
  for q in req.queries.iter() {
    
  }

  let mut resp = ReadResponse::new();

  // return to client.
  let resp_bytes_result = resp.write_to_bytes();
  if resp_bytes_result.is_err() {
    return Ok(create_error_response(500, resp_bytes_result.unwrap_err()));
  }
  let resp_bytes_compress_result = Encoder::new().compress_vec(&resp_bytes_result.unwrap());
  if resp_bytes_compress_result.is_err() {
    return Ok(create_error_response(500, resp_bytes_compress_result.unwrap_err()));
  }
  let resp_bytes = resp_bytes_compress_result.unwrap();
  let resp_body = warp::http::Response::builder()
    .status(200)
    .header("Content-Type", "")
    .body(Body::from(resp_bytes))
    .unwrap();
  Ok(resp_body)
}
