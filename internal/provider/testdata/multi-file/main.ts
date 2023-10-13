import { add } from "./util/calc.ts";
import operands from "./operands.json" with { type: "json" };

Deno.serve(() => {
  const sum = add(operands[0], operands[1]);
  return new Response(`sum: ${sum}`);
});
