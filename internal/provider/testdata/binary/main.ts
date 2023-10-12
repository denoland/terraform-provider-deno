import { join, dirname } from "https://deno.land/std@0.203.0/url/mod.ts";


Deno.serve(async () => {
    const url = join(dirname(import.meta.url), "computer_screen_programming.png");
    try {
        const image = await Deno.readFile(url);
        return new Response(image);
    } catch (error) {
        return new Response(error.message);
    }
});
