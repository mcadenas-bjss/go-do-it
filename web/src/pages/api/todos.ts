import type { APIRoute } from "astro";
import { BASE_URL, TODOS_PATH } from "../../utils/globals";

export async function GET() {
  try {
    const response = await fetch(`${BASE_URL}${TODOS_PATH}`);
    const data: Todos = await response.json();
    return new Response(JSON.stringify(data), { status: 200 })
  } catch (e) {
    return new Response(
      JSON.stringify({
        message: "An error occurred.",
      }),
      {
        status: 500,
      }
    );
  }
}