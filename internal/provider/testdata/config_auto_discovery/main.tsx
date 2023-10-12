/** @jsx h */

import { h } from "preact";
import { renderToString } from "preact-render-to-string";

Deno.serve((_req) => {
  const body = renderToString(<h1>Hello World!</h1>);
  return new Response(body, {
    headers: { "content-type": "text/html" },
  });
});
