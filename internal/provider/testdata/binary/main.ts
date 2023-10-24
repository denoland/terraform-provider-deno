Deno.serve(async () => {
  try {
    const image = await Deno.readFile("computer_screen_programming.png");
    return new Response(image);
  } catch (error) {
    return new Response(error.message);
  }
});
