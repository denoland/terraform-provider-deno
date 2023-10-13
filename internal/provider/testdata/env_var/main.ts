Deno.serve(() => new Response(`Hello ${Deno.env.get("FOO")}`));
