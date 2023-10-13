import { assert } from "std/assert/mod.ts";

assert(true);

Deno.serve(() => {
  return new Response("Hello World");
});
