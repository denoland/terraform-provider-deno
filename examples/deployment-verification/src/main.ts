import { walkSync } from "https://deno.land/std@0.200.0/fs/walk.ts";

const assets = [...walkSync("/")].map((entry) => ({
  ...entry,
  size: Deno.statSync(entry.path).size,
}));
const envVars = Object.fromEntries(
  Object.entries(Deno.env.toObject()).filter(
    ([key, _value]) => !key.startsWith("DENO_"),
  ),
);
Deno.serve(async () => {
  return new Response(
    await Deno.readFile("./src/images/computer_screen_programming.png"),
  );
  //return new Response(JSON.stringify({ assets, envVars }));
});
