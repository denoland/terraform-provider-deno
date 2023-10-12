import * as cowsay from "npm:cowsay";

Deno.serve(() => {
  const output = cowsay.say({ text: "Hello" });
  return new Response(output);
});
