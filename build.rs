/* coding: utf-8 */
/******************************************************************************
 * prometheus_sql_adapter
 *
 * Copyright 2020-, Kaede Fujisaki
 *****************************************************************************/
extern crate protobuf_codegen_pure;

fn main() {
  protobuf_codegen_pure::Codegen::new()
    .out_dir("src/proto")
    .inputs(&[
      "src/proto/remote.proto",
      "src/proto/types.proto",
    ])
    .include("src/proto")
    .run()
    .expect("protoc");
}
