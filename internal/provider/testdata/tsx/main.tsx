/** @jsx h */

import { h } from "npm:preact@10";
import { renderToString } from "npm:preact-render-to-string@6";

Deno.serve((_req) => {
  const body = renderToString(<h1>Hello World!</h1>);
  return new Response(body, {
    headers: { "content-type": "text/html" },
  });
});
