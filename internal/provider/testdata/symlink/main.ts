import { add } from "./symlink.js";

Deno.serve(() => {
  const sum = add(40, 2);
  return new Response(`sum: ${sum}`);
});
